// Package apitest は、Presentation の結合テストで使う共通の部品（テスト専用。本番コードからは import しない）。
// cmd と同じ部品（本物の spec の入力検証・operation ごとの認証・生成されたルート登録）でサーバーを組み立て、
// すべてのレスポンスを spec で検証する。
package apitest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

// Unimplemented は、テストで使わない operation を埋めるための入れ物。
// テスト用のサーバーは `struct{ apitest.Unimplemented; *xxx.Handler }` とし、Handler のメソッド（浅い階層）が優先される。
// 生成された IF を埋め込むので、使っていない operation を呼ぶと nil の呼び出しでテストが落ちる（気づける）。
type Unimplemented struct {
	webapi.ServerInterface
}

// Session は、テストのセッション Cookie の値。AcceptSession がこの値だけを有効とみなす。
const Session = "test-session"

// AcceptSession は、Session の Cookie だけを有効とみなすセッションの確認（実際のセッションはログインの結合テストで検証する）。
type AcceptSession struct{}

func (AcceptSession) Execute(_ context.Context, sessionID string) error {
	if sessionID != Session {
		return domainerr.New(domainerr.ErrUnauthorized, "unauthenticated", "ログインしてください")
	}
	return nil
}

// Server は、テスト用に組み立てたサーバーと、レスポンスの検証に使う spec。
type Server struct {
	e      *echo.Echo
	router routers.Router
}

// SessionChecker は、セッションを確かめる UseCase（authqry.CheckSessionUsecase と同じ形）。
type SessionChecker interface {
	Execute(ctx context.Context, sessionID string) error
}

// NewServer は、si の operation を本物の spec・入力検証・認証の middleware で登録したサーバーを作る。
func NewServer(t *testing.T, si webapi.ServerInterface, sessions SessionChecker) *Server {
	t.Helper()
	silenceLog(t)
	spec := loadSpec(t)
	validate, err := httpserver.ValidateRequest(spec)
	if err != nil {
		t.Fatal(err)
	}
	e := httpserver.NewEcho()
	webapi.RegisterHandlersWithOptions(e, si, webapi.RegisterHandlersOptions{
		OperationMiddlewares: httpserver.OperationMiddlewares(spec, httpserver.RequireSameOrigin(), httpserver.RequireAdminSession(sessions), validate),
	})
	router, err := gorillamux.NewRouter(loadSpec(t))
	if err != nil {
		t.Fatal(err)
	}
	return &Server{e: e, router: router}
}

// Request は、テストで送るリクエストの組み立て方。
type Request struct {
	Method string
	Path   string
	Body   string
	// Cookie はセッション Cookie の値（空なら付けない）。
	Cookie string
	// Origin は Origin ヘッダー（空なら自分と同じ http://example.com）。
	Origin string
	// Cancelled は、送る前に context を取り消す（DB の呼び出しを失敗させて 500 を作る）。
	Cancelled bool
}

// Do は、リクエストを送り、レスポンスを spec で検証してから返す（spec に無いステータス・形の違うボディは失敗）。
func (s *Server) Do(t *testing.T, r Request) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(r.Method, r.Path, strings.NewReader(r.Body))
	if r.Body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	origin := r.Origin
	if origin == "" {
		origin = "http://example.com"
	}
	req.Header.Set("Origin", origin)
	if r.Cookie != "" {
		req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: r.Cookie}) //nolint:gosec // テストのリクエストに付ける Cookie で、ブラウザに返すものではない
	}
	if r.Cancelled {
		ctx, cancel := context.WithCancel(req.Context())
		cancel()
		req = req.WithContext(ctx)
	}
	rec := httptest.NewRecorder()
	s.e.ServeHTTP(rec, req)
	s.validateResponse(t, r, rec)
	return rec
}

func (s *Server) validateResponse(t *testing.T, r Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	// 検証用のリクエストは本文を読まないので、元の本文は付けない
	req := httptest.NewRequest(r.Method, r.Path, nil)
	route, params, err := s.router.FindRoute(req)
	if err != nil {
		t.Fatalf("%s %s is not in the spec: %v", r.Method, r.Path, err)
	}
	err = openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route},
		Status:                 rec.Code,
		Header:                 rec.Header(),
		Body:                   io.NopCloser(strings.NewReader(rec.Body.String())),
		Options:                &openapi3filter.Options{IncludeResponseStatus: true, MultiError: true},
	})
	if err != nil {
		t.Fatalf("response of %s %s does not match the spec (status %d, body %s): %v", r.Method, r.Path, rec.Code, rec.Body.String(), err)
	}
}

// Decode は、レスポンスのボディを T に読む。
func Decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return v
}

// AssertStatus は、ステータスが want であることを確かめる。
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, want, rec.Body.String())
	}
}

// AssertCode は、ステータスと業務ルールの違反のコードを確かめる。
func AssertCode(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	AssertStatus(t, rec, status)
	if got := Decode[webapi.ErrorResponse](t, rec); got.Code == nil || *got.Code != code {
		t.Fatalf("code = %v, want %s (body=%s)", got.Code, code, rec.Body.String())
	}
}

// AssertFieldErrors は、400 と errors[{field, rule}] の中身を順不同で確かめる。
func AssertFieldErrors(t *testing.T, rec *httptest.ResponseRecorder, want ...webapi.FieldError) {
	t.Helper()
	AssertStatus(t, rec, http.StatusBadRequest)
	got := map[[2]string]bool{}
	if body := Decode[webapi.ErrorResponse](t, rec); body.Errors != nil {
		for _, e := range *body.Errors {
			got[[2]string{e.Field, e.Rule}] = true
		}
	}
	wantSet := map[[2]string]bool{}
	for _, e := range want {
		wantSet[[2]string{e.Field, e.Rule}] = true
	}
	if !reflect.DeepEqual(got, wantSet) {
		t.Fatalf("errors = %v, want %v (body=%s)", got, wantSet, rec.Body.String())
	}
}

// AssertNoFieldError は、field について rule の違反が返っていないことを確かめる（制約の「通る値」の確認に使う。
// 他の規則、例えば相関ルールの違反は、通る値の確認とは関係ないので見ない）。
func AssertNoFieldError(t *testing.T, rec *httptest.ResponseRecorder, field, rule string) {
	t.Helper()
	if rec.Code != http.StatusBadRequest {
		return
	}
	body := Decode[webapi.ErrorResponse](t, rec)
	if body.Errors == nil {
		return
	}
	for _, e := range *body.Errors {
		if e.Field == field && e.Rule == rule {
			t.Fatalf("unexpected violation %s/%s (body=%s)", e.Field, e.Rule, rec.Body.String())
		}
	}
}

// FieldError は、期待する違反を短く書くためのもの。
func FieldError(field, rule string) webapi.FieldError {
	return webapi.FieldError{Field: field, Rule: rule}
}

// loadSpec は生成物に埋め込んだ spec を読む（ValidateRequest が servers を書き換えるので、使うたびに読み直す）。
func loadSpec(t *testing.T) *openapi3.T {
	t.Helper()
	spec, err := webapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	spec.Servers = nil
	return spec
}

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}
