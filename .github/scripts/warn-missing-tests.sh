#!/usr/bin/env bash
# PR で変更したソースのうち、同じディレクトリのテストが変わっていないものを警告する（失敗にはしない）。
# リファクタリング等でテスト変更が不要なこともあるため、気づくための注意喚起に留める。
#
# 対象: backend/{internal,cmd}/**/*.go、frontend/*/src/**/*.ts(x)、frontend/packages/*/src/**/*.ts(x)
# 対象外: テスト自体・Story・型だけのファイル（*.d.ts）・テスト用の補助（frontend/*/src/testing）・
#         生成物（backend/api/gen・openapi/dist・frontend/*/src/api/generated）
# 結果は GitHub の注釈（1 ステップ 10 件まで表示）と、全件をジョブのサマリーに出す。
#
# 使い方: .github/scripts/warn-missing-tests.sh <比較元の ref>   （例: origin/develop）
set -euo pipefail

base=${1:?比較元の ref を指定してください（例: origin/develop）}
root=$(git rev-parse --show-toplevel)
cd "$root"

if ! git rev-parse --verify --quiet "${base}^{commit}" >/dev/null; then
  echo "::error title=比較元を解決できない::${base} を解決できません（actions/checkout の fetch-depth: 0 か、git fetch origin が必要です）"
  exit 1
fi

# GitHub の注釈のプロパティ値に使えない文字を逃がす
esc_prop() {
  local s=$1
  s=${s//%/%25}
  s=${s//$'\r'/%0D}
  s=${s//$'\n'/%0A}
  s=${s//:/%3A}
  s=${s//,/%2C}
  printf '%s' "$s"
}

# GitHub の注釈のメッセージ本文に使えない文字を逃がす
esc_data() {
  local s=$1
  s=${s//%/%25}
  s=${s//$'\r'/%0D}
  s=${s//$'\n'/%0A}
  printf '%s' "$s"
}

# 追加・変更に加え、内容を変えたリネーム（R）も対象にする（名前は変更後）。
# 日本語などのファイル名が引用・エスケープされないよう、-z（NUL 区切り）で読む
# プロセス置換では git diff の失敗（共通祖先が無い浅いクローン等）が伝わらず「変更なし」になるため、
# 一度ファイルに書いて終了ステータスを確かめる
diff_out=$(mktemp)
trap 'rm -f "$diff_out"' EXIT
if ! git diff -z --name-only --diff-filter=AMR "$base"...HEAD >"$diff_out"; then
  echo "::error title=差分を取れない::${base} と HEAD の差分を取れません（共通祖先が無い可能性があります。actions/checkout の fetch-depth: 0 を確認してください）"
  exit 1
fi
changed=()
while IFS= read -r -d '' f; do
  changed+=("$f")
done <"$diff_out"

# テストを追加・変更したディレクトリ（文字列の完全一致で比べる）
test_dirs=()
for f in "${changed[@]+"${changed[@]}"}"; do
  case "$f" in
    *_test.go | *.test.ts | *.test.tsx) test_dirs+=("$(dirname "$f")") ;;
  esac
done

has_test_change() {
  local d
  for d in "${test_dirs[@]+"${test_dirs[@]}"}"; do
    [ "$d" = "$1" ] && return 0
  done
  return 1
}

missing=()
for f in "${changed[@]+"${changed[@]}"}"; do
  case "$f" in
    *_test.go | *.test.ts | *.test.tsx | *.stories.ts | *.stories.tsx | *.d.ts) continue ;;
    */generated/* | frontend/*/src/testing/*) continue ;;
    # case の * は / にも一致するので、frontend/*/src は frontend/packages/ui/src も含む
    backend/internal/*.go | backend/cmd/*.go) ;;
    frontend/*/src/*.ts | frontend/*/src/*.tsx) ;;
    *) continue ;;
  esac
  if ! has_test_change "$(dirname "$f")"; then
    missing+=("$f")
    echo "::warning file=$(esc_prop "$f"),title=テストの変更が無い::$(esc_data "$f") を変更したが、同じディレクトリのテストが追加・変更されていません。振る舞いが変わるならテストを追加してください（docs/rules/testing.md）"
  fi
done

echo "テストの変更が無いソース: ${#missing[@]} 件（警告のみ）"

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
  {
    echo "## テストの変更が無いソース（警告のみ）: ${#missing[@]} 件"
    echo
    if [ ${#missing[@]} -gt 0 ]; then
      echo "振る舞いが変わる変更なら、同じディレクトリにテストを追加してください（docs/rules/testing.md）。"
      echo
      for f in "${missing[@]}"; do echo "- \`${f}\`"; done
    fi
  } >>"$GITHUB_STEP_SUMMARY"
fi
