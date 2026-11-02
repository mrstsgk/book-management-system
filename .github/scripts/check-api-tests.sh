#!/usr/bin/env bash
# API の各 operation に、生成物を使う側のテストがあるかを検査する（docs/rules/testing.md「テストの有無の検査」）。
# 生成物（backend/api/docs・frontend/*/src/api/generated）はテストもレビューもしないが、
# それを実装・利用するコードのテストは必須。
#
# operation の一覧は swag が排出した backend/api/docs/swagger.json の paths から取る
# （BE の正は Go の Handler で、swagger.json は CI のドリフト検査で Handler と一致している）。
#
# 判定（テスト・ソースはコメントを除いて見る）:
#   BE: backend/internal/presentation/http/ 配下の *_test.go のいずれかに、そのパスの文字列（{id} は任意の 1 セグメント）と
#       HTTP メソッド（http.MethodGet または "GET"）の両方があること
#   FE: アプリのコード（生成物・テスト・Story・src/testing 以外）が生成モジュール api/generated/api から
#       この operation の hook・関数（use<Op> / use<Op>Infinite / get<Op>QueryOptions / get<Op>QueryKey / <op>）を
#       使っていたら、*.test.ts(x) のいずれかが生成 MSW ハンドラ get<Op>MockHandler( を呼ぶか、
#       同じメソッド・パスの手書き MSW ハンドラ（http.put('*/api/tags/:id' …)）を使っていること
#       （生成 API の差し替えは HTTP 境界（MSW）で行い vi.mock() しない規約。まだ使っていない operation は対象外）。
#       <Op> は Orval の命名（operationId が無いので HTTP メソッド + パス。GET /api/books/{id} → GetApiBooksId）
#
# 限界: 文字列での判定なので、同名の別 operation のテストに一致しうる。
#       コメントの除去は文字列リテラルを解釈しないため、文字列中の /* … */ や // もコメントとして除く
#       （"http://" のように : の直後の // は残す）。
#
# 使い方: .github/scripts/check-api-tests.sh   （リポジトリ内ならどこからでも。python3 と perl が要る）
set -euo pipefail

root=$(git rev-parse --show-toplevel)
fail=0

spec="$root/backend/api/docs/swagger.json"
if [ ! -f "$spec" ]; then
  echo "::error title=API 契約がない::backend/api/docs/swagger.json がありません（make -C backend swagger で排出してください）"
  exit 1
fi

# ファイルの内容からコメント（/* */ と //）を除いて出す。"http://" のように : の直後の // は残す
strip_comments() {
  perl -0777 -pe 's{/\*.*?\*/}{}gs; s{(?<!:)//[^\n]*}{}g' "$1"
}

# いずれかのファイル（コメントを除く）が正規表現に一致するか。引数: 正規表現 ファイル...
# grep -q は一致した時点で読むのをやめ、pipefail 下では前段が SIGPIPE で失敗して不一致扱いになるため使わない
any_match() {
  local re=$1 f
  shift
  for f in "$@"; do
    [ -f "$f" ] || continue
    if strip_comments "$f" | grep -E "$re" >/dev/null; then
      return 0
    fi
  done
  return 1
}

# いずれかのファイル（コメントを除く）が 2 つの正規表現の両方に一致するか。引数: 正規表現1 正規表現2 ファイル...
any_match_both() {
  local re1=$1 re2=$2 f body
  shift 2
  for f in "$@"; do
    [ -f "$f" ] || continue
    body=$(strip_comments "$f")
    if printf '%s\n' "$body" | grep -E "$re1" >/dev/null && printf '%s\n' "$body" | grep -E "$re2" >/dev/null; then
      return 0
    fi
  done
  return 1
}

# 正規表現の特殊文字を文字として扱う
re_escape() {
  printf '%s' "$1" | sed -e 's/[.[\*^$+?(){}|/]/\\&/g'
}

# swagger.json の paths から「HTTP メソッド パス」を 1 行ずつ出す（メソッドは大文字）
operations() {
  python3 - "$spec" <<'PY'
import json, sys
spec = json.load(open(sys.argv[1]))
for path, methods in spec.get("paths", {}).items():
    for method in methods:
        if method.lower() in ("get", "post", "put", "patch", "delete", "head", "options"):
            print(method.upper(), path)
PY
}

# Orval の名前（HTTP メソッド + パス）。/api/books/{id} → ApiBooksId
op_name() {
  local path=$1 seg out=""
  IFS='/' read -r -a segs <<<"$path"
  for seg in "${segs[@]}"; do
    [ -n "$seg" ] || continue
    seg=${seg#\{}
    seg=${seg%\}}
    # kebab-case の区切りも大文字にする（tag-counts → TagCounts）
    seg=$(printf '%s' "$seg" | perl -pe 's/(^|-)([a-z0-9])/\U$2/g')
    out+="$seg"
  done
  printf '%s' "$out"
}

be_tests=()
while IFS= read -r f; do be_tests+=("$f"); done < <(
  find "$root/backend/internal/presentation/http" -type f -name '*_test.go' 2>/dev/null
)

app_src="$root/frontend/web/src"
fe_gen="$app_src/api/generated"
if [ -d "$app_src" ] && [ ! -d "$fe_gen" ]; then
  echo "::error title=FE の生成物がない::frontend/web/src/api/generated がありません（pnpm -C frontend gen:api で生成してください）"
  fail=1
fi
fe_tests=()
fe_sources=()
if [ -d "$app_src" ]; then
  while IFS= read -r f; do fe_tests+=("$f"); done < <(
    find "$app_src" -type f \( -name '*.test.ts' -o -name '*.test.tsx' \) -not -path '*/generated/*'
  )
  while IFS= read -r f; do fe_sources+=("$f"); done < <(
    find "$app_src" -type f \( -name '*.ts' -o -name '*.tsx' \) \
      -not -name '*.test.ts' -not -name '*.test.tsx' -not -name '*.stories.ts' -not -name '*.stories.tsx' \
      -not -path '*/generated/*' -not -path '*/testing/*'
  )
fi

n=0
while read -r http_method path; do
  [ -n "$path" ] || continue
  n=$((n + 1))
  # http.MethodGet の Get（GET → Get）
  method_word="${http_method:0:1}$(printf '%s' "${http_method:1}" | tr '[:upper:]' '[:lower:]')"
  # パスパラメータ（{id}）は任意の 1 セグメント
  path_re=$(re_escape "$path" | sed -e 's/\\{[A-Za-z0-9_]*\\}/[^\/"]+/g')

  # ---- BE ----
  # パスの文字列（{id} は任意の 1 セグメント）と HTTP メソッド（http.MethodGet か "GET"）が同じテストファイルにあること。
  # 同じ行に限定しないのは、リクエストを組み立てる補助関数（serve(t, h, path, ...) 等）にメソッドが隠れるため
  if any_match_both "\"(https?://[^/\"]*)?${path_re}[\"?]" "(http\.Method${method_word}[^A-Za-z0-9_]|\"${http_method}\")" "${be_tests[@]+"${be_tests[@]}"}"; then
    echo "ok  BE ${http_method} ${path}"
  else
    echo "::error title=BE のテストがない::${http_method} ${path} のテストが backend/internal/presentation/http/ の *_test.go にありません。http.Method${method_word} で ${path} にリクエストするテストを追加してください"
    fail=1
  fi

  # ---- FE ----
  [ -d "$fe_gen" ] || continue
  Op="${method_word}$(op_name "$path")"
  op="$(printf '%s' "${Op:0:1}" | tr '[:upper:]' '[:lower:]')${Op:1}"
  # Orval の版や設定で export function / export const のどちらにもなりうる
  if ! grep -qE "export (function|const) use${Op}[[:space:]<(=:]" "$fe_gen/api.ts" 2>/dev/null ||
    ! grep -qE "export (function|const) get${Op}MockHandler[[:space:]<(=:]" "$fe_gen/api.msw.ts" 2>/dev/null; then
    echo "::error title=FE の生成物の名前を解決できない::${http_method} ${path} の use${Op} / get${Op}MockHandler が frontend/web/src/api/generated にありません（生成物が古ければ pnpm -C frontend gen:api。Orval の命名が変わったなら本スクリプトの op_name() を直してください）"
    fail=1
    continue
  fi
  used=""
  for f in "${fe_sources[@]+"${fe_sources[@]}"}"; do
    body=$(strip_comments "$f")
    if printf '%s\n' "$body" | grep -E "api/generated/api['\"]" >/dev/null &&
      printf '%s\n' "$body" | grep -E "(^|[^A-Za-z0-9_])(use${Op}(Infinite)?|get${Op}(Infinite)?Query(Options|Key)|${op})([^A-Za-z0-9_]|$)" >/dev/null; then
      used+="${f#"$root"/} "
    fi
  done
  if [ -z "$used" ]; then
    echo "--  FE ${http_method} ${path}（アプリのコードでまだ使っていないため対象外）"
    continue
  fi
  # 生成ハンドラか、同じメソッド・パスの手書き MSW ハンドラ（http.put('*/api/tags/:id' …)。4xx やリクエスト内容の検証に使う）
  method_lower=$(printf '%s' "$http_method" | tr '[:upper:]' '[:lower:]')
  msw_path_re=$(re_escape "$path" | perl -pe 's/\\\{[A-Za-z0-9_]*\\\}/(:[A-Za-z0-9_]+|[^\/\x27"]+)/g')
  if any_match "get${Op}MockHandler\(" "${fe_tests[@]+"${fe_tests[@]}"}" ||
    any_match "http\.${method_lower}\(['\"]\*?${msw_path_re}['\"]" "${fe_tests[@]+"${fe_tests[@]}"}"; then
    echo "ok  FE ${http_method} ${path}"
  else
    echo "::error title=FE のテストがない::${http_method} ${path} を使うコード（${used% }）に対し、生成 MSW ハンドラ get${Op}MockHandler( か、同じメソッド・パスの MSW ハンドラ（http.${method_lower}('*${path}' …)）を使うテストが frontend/web/src にありません"
    fail=1
  fi
done < <(operations)

if [ "$n" -eq 0 ]; then
  echo "::error title=operation を読めない::backend/api/docs/swagger.json から operation を 1 件も読めませんでした"
  fail=1
fi

exit "$fail"
