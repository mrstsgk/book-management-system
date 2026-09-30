#!/usr/bin/env bash
#
# Stop hook: 変更に対するテスト追加チェック
# ------------------------------------------
# develop からの分岐後に変更されたファイル（コミット済み+未コミット+未追跡）を見て、
# 各スタックのソースを変更したのに対応するテストが無ければ Stop をブロックする。
#
# 「対応するテスト」は同じディレクトリ・同名の *_test.go / *.test.ts(x)（testing.md の
# コロケーション規約）に厳密一致させる。パス全体の部分一致にすると、同じ親ディレクトリ名
# （例: frontend の src/ 直下や components/）を含むだけの無関係なテスト変更で
# 通過してしまうため（ソースと同名でない別機能のテストが誤って一致する）。
#
# 削除されたファイルは対象外。消えたコードに対してテストは書けず、削除の妥当性は
# 「残ったコードが壊れていないか」を既存テストで確認する形でしか担保できないため
# （対象外にしないと、空ファイルや未使用コードの掃除で必ずブロックされる）。
#
# 注意: Stop hook はブロックするたびにモデルを再呼び出しする。指摘に応えられない
# 変更（例: 生成物だけの更新）で引っかかると同じ指摘が繰り返し出続けるため、
# そのときは SKIP_TEST_CHECK=1 で抜ける。
#
# 対象:
#   backend   backend/{internal,cmd}/**/*.go  （*_test.go を除く）
#   frontend  frontend/{web,packages/ui}/src/**/*.{ts,tsx}
#             （*.test.* / *.stories.* / *.d.ts / src/testing/ / 生成物 src/api/generated/・src/types/api.ts を除く）
#
# swag 排出物（backend/api/docs/）は internal/cmd の外なので対象にならない。
#
# バイパス: SKIP_TEST_CHECK=1

command -v jq >/dev/null 2>&1 || exit 0

case "${SKIP_TEST_CHECK:-}" in
    1|true|TRUE|yes) exit 0 ;;
esac

top="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
cd "$top" || exit 0

branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
case "$branch" in
    main|develop) exit 0 ;;  # 永続ブランチはチェック対象外
esac

merge_base="$(git merge-base origin/develop HEAD 2>/dev/null || git merge-base develop HEAD 2>/dev/null)"
[ -n "$merge_base" ] || exit 0

# 追跡済みファイル（コミット済み+未コミットのステージ済み/未ステージの変更）は
# git diff だけで拾える。--diff-filter=d で削除を除外する。
# 未追跡ファイルは git ls-files --others で別途拾う。
#
# git status --porcelain の $2 を空白区切り（awk）で取り出す実装は使わない。
# リネーム時は "R  old -> new" の旧パス側を拾ってしまい、パスに空白を含む
# 場合は途中で切り詰められるため。
changed="$(
    { git diff --name-only --diff-filter=d "$merge_base" 2>/dev/null
      git ls-files --others --exclude-standard 2>/dev/null
    } | sort -u
)"
[ -n "$changed" ] || exit 0

# $1 = 変更ファイル一覧の名前(未使用) → jq block を出力して抜ける
_report() {
    local name="$1" uncovered="$2"
    local files
    files="$(printf '%s' "$uncovered" | sed 's/^/  - /' | head -n 20)"
    jq -n --arg r "${name}: 以下のソースを変更しましたが、同じディレクトリ・同名の対応テストの変更が見当たりません。変更箇所とその影響範囲に対するテストを追加してください（ルール: docs/rules/testing.md。意図的にスキップする場合は SKIP_TEST_CHECK=1）。

対応するテストが見つからない変更ファイル:
${files}" '{decision:"block", reason:$r}'
    exit 0
}

# backend: foo.go → 同じディレクトリの foo_test.go（Go のコロケーション規約）
_check_backend() {
    local src_files uncovered="" f expected
    src_files="$(printf '%s\n' "$changed" | grep -E '^backend/(internal|cmd)/.*\.go$' | grep -vE '_test\.go$' || true)"
    [ -n "$src_files" ] || return 0

    while IFS= read -r f; do
        [ -n "$f" ] || continue
        expected="${f%.go}_test.go"
        printf '%s\n' "$changed" | grep -qFx "$expected" || uncovered="${uncovered}${f}"$'\n'
    done <<< "$src_files"

    [ -n "$uncovered" ] || return 0
    _report "backend" "$uncovered"
}

# frontend: Foo.ts(x) → 同じディレクトリの Foo.test.ts または Foo.test.tsx（bulletproof のコロケーション規約）
_check_frontend() {
    local src_files uncovered="" f base
    src_files="$(printf '%s\n' "$changed" \
        | grep -E '^frontend/(web|packages/ui)/src/.*\.(ts|tsx)$' \
        | grep -vE '\.test\.(ts|tsx)$|\.stories\.(ts|tsx)$|\.d\.ts$|/src/testing/|/src/api/generated/|/src/types/api\.ts$' || true)"
    [ -n "$src_files" ] || return 0

    # 実装が .ts でもテストは .test.tsx（renderHook で wrapper を使う hooks 等）になりうるため、どちらの拡張子でも対応とみなす
    while IFS= read -r f; do
        [ -n "$f" ] || continue
        base="${f%.*}"
        printf '%s\n' "$changed" | grep -qFx -e "${base}.test.ts" -e "${base}.test.tsx" || uncovered="${uncovered}${f}"$'\n'
    done <<< "$src_files"

    [ -n "$uncovered" ] || return 0
    _report "frontend" "$uncovered"
}

_check_backend
_check_frontend

exit 0
