package httpserver

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// adminPathPrefix は管理（自分だけ）の API の path。Origin の検証を掛ける範囲。
const adminPathPrefix = "/api/admin/"

// sessionSchemeName は spec の securitySchemes に書いたセッション Cookie の名前。
const sessionSchemeName = "adminSession"

// OperationMiddlewares は、operation ごとに掛ける middleware を spec から決める（生成コードの RegisterHandlersOptions に渡す）。
// 順は「Origin の検証 → セッションの検証 → 入力検証」。未ログインの要求に、入力の形の情報（400 の errors）を返さないため。
//   - Origin の検証: 管理（/api/admin）の状態を変える要求（GET 以外。ログイン・ログアウトを含む）
//   - セッションの検証: spec の security に adminSession がある operation
//   - 入力検証: すべての operation
//
// 一覧を手で書かずに spec から作るのは、operation を足したときに認証の掛け忘れが起きないようにするため。
func OperationMiddlewares(spec *openapi3.T, sameOrigin, session, validate echo.MiddlewareFunc) map[string][]echo.MiddlewareFunc {
	out := map[string][]echo.MiddlewareFunc{}
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			var mws []echo.MiddlewareFunc
			if strings.HasPrefix(path, adminPathPrefix) && method != http.MethodGet {
				mws = append(mws, sameOrigin)
			}
			if requiresSession(op) {
				mws = append(mws, session)
			}
			out[operationKey(op.OperationID)] = append(mws, validate)
		}
	}
	return out
}

func requiresSession(op *openapi3.Operation) bool {
	if op.Security == nil {
		return false
	}
	for _, req := range *op.Security {
		if _, ok := req[sessionSchemeName]; ok {
			return true
		}
	}
	return false
}

// operationKey は、生成コードの OperationMiddlewares の鍵（YAML に書いた lowerCamel の operationId）に合わせる。
// 生成物に埋め込まれた spec の operationId は Go の名前に正規化されて先頭が大文字になっているため（searchBooks → SearchBooks）。
func operationKey(id string) string {
	if id == "" {
		return id
	}
	return strings.ToLower(id[:1]) + id[1:]
}
