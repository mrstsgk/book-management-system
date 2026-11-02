#!/usr/bin/env bash
# API の各 operation に、生成物を使う側のテストがあるかを検査する（docs/rules/testing.md「テストの有無の検査」）。
# 生成物（backend/api/gen・frontend/*/src/api/generated）はテストもレビューもしないが、
# それを実装・利用するコードのテストは必須。
#
# バンドル（チャネル。実行サービス cmd/<channel>-api ごと）は openapi/dist/<bundle>.yaml から取り、operation の一覧と名前は推測せず生成物から読む
# （CI の生成物ドリフト検査で openapi/dist と一致している）。
#   - BE: backend/api/gen/<bundle>/api.gen.go のルート登録
#         router.GET(options.BaseURL+"/api/books", wrapper.SearchBooks, options.OperationMiddlewares["searchBooks"]...)
#   - FE: frontend/<app>/src/api/generated/{api,api.msw}.ts の use<Op> / get<Op>MockHandler
#
# 判定（テスト・ソースはコメントを除いて見る）:
#   BE: backend/internal/*/presentation/http/*/ 配下（チャネル public / admin）のテスト以外の *.go にその operation の
#       Handler のメソッド（func (…) SearchBooks(）があれば、同じ配下の *_test.go のいずれかに、
#       その HTTP メソッドとパスへのリクエスト（http.MethodGet または "GET" と "/api/books" が同じ行。
#       {id} は任意の 1 セグメント）があること
#       （Handler がまだ無い operation は対象外。詳細設計の PR で契約と生成物だけを先にマージし、BE・FE を並行して実装するため。
#       docs/rules/development-flow.md §3。cmd に置く未実装のスタブは presentation の外なので Handler と数えない）
#   FE: アプリのコード（生成物・テスト・Story・src/testing 以外）が生成モジュール api/generated/api から
#       この operation の hook・関数（use<Op> / use<Op>Infinite / get<Op>QueryOptions / get<Op>QueryKey / <op>）を
#       使っていたら、*.test.ts(x) のいずれかが生成 MSW ハンドラ get<Op>MockHandler( を呼ぶか、
#       同じメソッド・パスの手書き MSW ハンドラ（http.put('*/api/admin/tags/:id' …)。4xx やリクエスト内容の検証に使う）を
#       使っていること（生成 API の差し替えは MSW で行い vi.mock() しない規約。まだ使っていない operation は対象外）
#
# 限界: 文字列での判定なので、同じパスを別の operation のテストが使っていると一致する。
#       コメントの除去は文字列リテラルを解釈しないため、文字列中の /* … */ や // もコメントとして除く
#       （"http://" のように : の直後の // は残す）。
#
# 使い方: .github/scripts/check-api-tests.sh   （リポジトリ内ならどこからでも）
set -euo pipefail

root=$(git rev-parse --show-toplevel)
fail=0

# バンドル（チャネル）名 → フロントエンドのアプリのディレクトリ名（public → frontend/public、admin → frontend/admin。同じ名前）
app_of() {
  echo "$1"
}

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

# 正規表現の特殊文字を文字として扱う
re_escape() {
  printf '%s' "$1" | sed -e 's/[.[\*^$+?(){}|/]/\\&/g'
}

# 生成された Echo のルート登録から「HTTP メソッド パス Go のメソッド名 operationId」を 1 行ずつ出す
routes() {
  sed -n 's/.*router\.\([A-Z]*\)(options\.BaseURL+"\([^"]*\)", wrapper\.\([A-Za-z0-9_]*\), options\.OperationMiddlewares\["\([^"]*\)"\].*/\1 \2 \3 \4/p' "$1"
}

# バンドルの一覧は openapi/dist/<bundle>.yaml（openapi/redocly.yaml の apis ごと。OpenAPI CI の bundle-check で一致を保証）から取る。
# 既にある生成物から数えると、生成物が欠けたバンドルが検査から黙って外れるため
shopt -s nullglob
specs=("$root"/openapi/dist/*.yaml)
if [ ${#specs[@]} -eq 0 ]; then
  echo "::error title=API 契約がない::openapi/dist/*.yaml がありません（make -C openapi bundle でバンドルしてください）"
  exit 1
fi

be_tests=()
be_sources=()
while IFS= read -r f; do be_sources+=("$f"); done < <(
  find "$root/backend/internal" -type f -path "*/presentation/http/*" -name '*.go' -not -name '*_test.go' 2>/dev/null
)
while IFS= read -r f; do be_tests+=("$f"); done < <(
  find "$root/backend/internal" -type f -path "*/presentation/http/*" -name '*_test.go' 2>/dev/null
)

for spec in "${specs[@]}"; do
  bundle=$(basename "$spec" .yaml)
  gen="$root/backend/api/gen/${bundle}/api.gen.go"
  # 生成物が無いと何も検査せずに通ってしまうため、欠落は失敗にする
  if [ ! -f "$gen" ]; then
    echo "::error title=BE の生成物がない::${bundle} の backend/api/gen/${bundle}/api.gen.go がありません（make -C backend api-gen で生成してください）"
    fail=1
    continue
  fi
  app=$(app_of "$bundle")
  app_src="$root/frontend/$app/src"
  fe_gen="$app_src/api/generated"
  # FE のアプリがまだ無いバンドルは FE の検査をしない。アプリがあるのに生成物が無いのは失敗
  if [ -d "$app_src" ] && [ ! -d "$fe_gen" ]; then
    echo "::error title=FE の生成物がない::frontend/${app}/src/api/generated がありません（pnpm -C frontend gen:api で生成してください）"
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
  while read -r http_method path go_method op; do
    [ -n "$op" ] || continue
    n=$((n + 1))
    # http.MethodGet の Get（GET → Get）
    method_word="${http_method:0:1}$(printf '%s' "${http_method:1}" | tr '[:upper:]' '[:lower:]')"
    # Echo のパスパラメータ（:id）は任意の 1 セグメント
    path_re=$(re_escape "$path" | sed -e 's/:[A-Za-z0-9_]*/[^\/"]+/g')

    # ---- BE ----
    if ! any_match "func \([^)]*\) ${go_method}\(" "${be_sources[@]+"${be_sources[@]}"}"; then
      echo "--  BE ${bundle} ${op}（Handler がまだ無いため対象外）"
    elif any_match "(http\.Method${method_word}[^A-Za-z0-9_]|\"${http_method}\").*\"(https?://[^/\"]*)?${path_re}[\"?]" "${be_tests[@]+"${be_tests[@]}"}"; then
      echo "ok  BE ${bundle} ${op}"
    else
      echo "::error title=BE のテストがない::operation '${op}'（${http_method} ${path}）のテストが backend/internal/*/presentation/http/ の *_test.go にありません。http.Method${method_word} で ${path} にリクエストするテストを追加してください"
      fail=1
    fi

    # ---- FE ----
    [ -d "$fe_gen" ] || continue
    # Orval の名前は operationId の先頭を大文字にしたもの（operationId は lowerCamel の規約）
    Op="$(printf '%s' "${op:0:1}" | tr '[:lower:]' '[:upper:]')${op:1}"
    # Orval の版や設定で export function / export const のどちらにもなりうる
    if ! grep -qE "export (function|const) use${Op}[[:space:]<(=:]" "$fe_gen/api.ts" 2>/dev/null ||
      ! grep -qE "export (function|const) get${Op}MockHandler[[:space:]<(=:]" "$fe_gen/api.msw.ts" 2>/dev/null; then
      echo "::error title=FE の生成物の名前を解決できない::operation '${op}' の use${Op} / get${Op}MockHandler が frontend/${app}/src/api/generated にありません（operationId は lowerCamel にする。生成物が古ければ pnpm -C frontend gen:api）"
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
      echo "--  FE ${bundle} ${op}（アプリのコードでまだ使っていないため対象外）"
      continue
    fi
    # 生成ハンドラか、同じメソッド・パスの手書き MSW ハンドラ（http.put('*/api/admin/tags/:id' …)）
    method_lower=$(printf '%s' "$http_method" | tr '[:upper:]' '[:lower:]')
    msw_path_re=$(re_escape "$path" | sed -e "s/:[A-Za-z0-9_]*/[^\/'\"]+/g")
    if any_match "get${Op}MockHandler\(" "${fe_tests[@]+"${fe_tests[@]}"}" ||
      any_match "http\.${method_lower}\(['\"]\*?${msw_path_re}['\"]" "${fe_tests[@]+"${fe_tests[@]}"}"; then
      echo "ok  FE ${bundle} ${op}"
    else
      echo "::error title=FE のテストがない::operation '${op}' を使うコード（${used% }）に対し、生成 MSW ハンドラ get${Op}MockHandler( か、同じメソッド・パスの MSW ハンドラ（http.${method_lower}('*${path}' …)）を使うテストが frontend/${app}/src にありません"
      fail=1
    fi
  done < <(routes "$gen")

  if [ "$n" -eq 0 ]; then
    echo "::error title=ルート登録を読めない::${gen#"$root"/} から operation を 1 件も読めませんでした（生成物の形式が変わった可能性があります。本スクリプトの routes() を直してください）"
    fail=1
  fi
done

exit "$fail"
