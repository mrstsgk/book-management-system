package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

// testSpec は、API の YAML で使っている制約の種類を 1 つずつ持つテスト用の spec。
// YAML で新しい種類の制約を使うときは、ここに足して 400 への変換を確かめる（docs/rules/testing.md）。
const testSpec = `
openapi: 3.0.3
info: {title: test, version: "0"}
servers: [{url: /}]
paths:
  /items/{id}:
    post:
      operationId: registerItem
      parameters:
        - {name: id, in: path, required: true, schema: {type: integer, format: int64, minimum: 1}}
        - {name: limit, in: query, required: false, schema: {type: integer, minimum: 1, maximum: 100}}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [title, rating]
              properties:
                title: {type: string, minLength: 1, maxLength: 5}
                rating: {type: integer, minimum: 1, maximum: 5}
                tagIds: {type: array, maxItems: 2, items: {type: integer, format: int64}}
      responses:
        '204': {description: ok}
`

func newValidatedEcho(t *testing.T) *echo.Echo {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromData([]byte(testSpec))
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	mw, err := httpserver.ValidateRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	e := httpserver.NewEcho()
	e.Use(mw)
	// Handler が自分でボディを読むので、検証の後も読めることを確かめる
	e.POST("/items/:id", func(c echo.Context) error {
		var body map[string]any
		if err := c.Bind(&body); err != nil {
			return err
		}
		if _, ok := body["title"]; !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "body was consumed by the validator")
		}
		return c.NoContent(http.StatusNoContent)
	})
	return e
}

func TestValidateRequest(t *testing.T) {
	silenceLog(t)
	e := newValidatedEcho(t)

	tests := []struct {
		name        string
		path        string
		body        string
		wantStatus  int
		wantMessage string
		wantErrors  []httpserver.FieldError
	}{
		{name: "仕様どおりのリクエストはボディごと Handler に届く", path: "/items/1?limit=1", body: `{"title":"a","rating":1,"tagIds":[1,2]}`, wantStatus: http.StatusNoContent},
		{name: "必須のプロパティが無ければ required", path: "/items/1", body: `{"rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "title", Rule: "required"}}},
		{name: "最小長未満は minLength", path: "/items/1", body: `{"title":"","rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "title", Rule: "minLength"}}},
		{name: "最大長ちょうどは通る（日本語は 1 文字を 1 と数える）", path: "/items/1", body: `{"title":"あいうえお","rating":1}`, wantStatus: http.StatusNoContent},
		{name: "最大長+1 は maxLength", path: "/items/1", body: `{"title":"あいうえおか","rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "title", Rule: "maxLength"}}},
		{name: "最小値未満は minimum", path: "/items/1", body: `{"title":"a","rating":0}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "rating", Rule: "minimum"}}},
		{name: "最大値ちょうどは通る", path: "/items/1", body: `{"title":"a","rating":5}`, wantStatus: http.StatusNoContent},
		{name: "最大値+1 は maximum", path: "/items/1", body: `{"title":"a","rating":6}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "rating", Rule: "maximum"}}},
		{name: "配列の件数の上限+1 は maxItems", path: "/items/1", body: `{"title":"a","rating":1,"tagIds":[1,2,3]}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "tagIds", Rule: "maxItems"}}},
		{name: "配列の要素の型違いは添字付きのパス", path: "/items/1", body: `{"title":"a","rating":1,"tagIds":[1,"x"]}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "tagIds.1", Rule: "type"}}},
		{name: "型違いは type", path: "/items/1", body: `{"title":"a","rating":"1"}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "rating", Rule: "type"}}},
		{name: "独立した複数の違反はすべて返す", path: "/items/1", body: `{"title":"","rating":6}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request",
			wantErrors: []httpserver.FieldError{{Field: "rating", Rule: "maximum"}, {Field: "title", Rule: "minLength"}}},
		{name: "JSON として読めないボディは invalid json", path: "/items/1", body: `{"title":`, wantStatus: http.StatusBadRequest, wantMessage: "invalid json"},
		{name: "オブジェクトの代わりに配列を送ると errors の無い invalid request", path: "/items/1", body: `[1,2]`, wantStatus: http.StatusBadRequest, wantMessage: "invalid request"},
		{name: "必須のボディが無ければ invalid json", path: "/items/1", body: ``, wantStatus: http.StatusBadRequest, wantMessage: "invalid json"},
		{name: "path パラメータの最小値未満はパラメータ名で返す", path: "/items/0", body: `{"title":"a","rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "id", Rule: "minimum"}}},
		{name: "path パラメータの型違いは type", path: "/items/abc", body: `{"title":"a","rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "id", Rule: "type"}}},
		{name: "query パラメータの最大値+1 は maximum", path: "/items/1?limit=101", body: `{"title":"a","rating":1}`,
			wantStatus: http.StatusBadRequest, wantMessage: "invalid request", wantErrors: []httpserver.FieldError{{Field: "limit", Rule: "maximum"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusBadRequest {
				return
			}
			var got httpserver.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			if !reflect.DeepEqual(got.Errors, tt.wantErrors) {
				t.Errorf("errors = %+v, want %+v", got.Errors, tt.wantErrors)
			}
		})
	}
}

func TestValidateRequest_specに無いルートはそのまま通す(t *testing.T) {
	silenceLog(t)
	rec := httptest.NewRecorder()
	newValidatedEcho(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}
