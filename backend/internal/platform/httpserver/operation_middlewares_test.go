package httpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

const middlewareSpec = `
openapi: 3.0.3
info: {title: test, version: "0"}
components:
  securitySchemes:
    adminSession: {type: apiKey, in: cookie, name: admin_session}
paths:
  /api/items:
    get:
      operationId: searchItems
      security: []
      responses: {'200': {description: ok}}
  /api/admin/items:
    get:
      operationId: lookupItem
      security: [{adminSession: []}]
      responses: {'200': {description: ok}}
    post:
      operationId: registerItem
      security: [{adminSession: []}]
      responses: {'200': {description: ok}}
  /api/admin/login:
    post:
      operationId: login
      security: []
      responses: {'200': {description: ok}}
`

// TestOperationMiddlewares は、どの operation にどの検証が掛かるかを spec から決めていることを確かめる。
// 生成コードの登録の鍵（lowerCamel の operationId）で引けることも確かめる。
func TestOperationMiddlewares(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData([]byte(middlewareSpec))
	if err != nil {
		t.Fatal(err)
	}
	// 生成物に埋め込まれた spec と同じく、operationId の先頭が大文字に正規化されていても同じ鍵になること
	spec.Paths.Find("/api/admin/login").Post.OperationID = "Login"

	var calls []string
	mark := func(name string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { calls = append(calls, name); return next(c) }
		}
	}
	got := httpserver.OperationMiddlewares(spec, mark("origin"), mark("session"), mark("validate"))

	tests := []struct {
		op   string
		want string
	}{
		{"searchItems", "validate"},
		{"lookupItem", "session,validate"},
		{"registerItem", "origin,session,validate"},
		{"login", "origin,validate"},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			calls = nil
			h := echo.HandlerFunc(func(echo.Context) error { return nil })
			mws := got[tt.op]
			for _, mw := range slices.Backward(mws) {
				h = mw(h)
			}
			e := echo.New()
			_ = h(e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.Background()), httptest.NewRecorder()))
			if s := strings.Join(calls, ","); s != tt.want {
				t.Fatalf("middlewares = %q, want %q", s, tt.want)
			}
		})
	}
}
