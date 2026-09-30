package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	infraauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/auth"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/auth"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
	httpauth "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/query"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

// Presentation → Infrastructure の結合テスト（docs/rules/testing.md）。
// 実 DB のセッションと bcrypt を通して、login → 書き込み API → logout → 同じ Cookie で 401 の配線を確かめる。
// 分岐網羅は handler_test.go・usecase・契約テストが担う。

func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgcommon.Connect(pgcommon.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up migrate-up` first): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM admin_session LIMIT 0").Error; err != nil {
		t.Skipf("skipping: schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// newIntegrationEcho は cmd/api/main.go と同じ形で auth と tag を配線する。bcrypt はテストが速いよう最小コスト。
func newIntegrationEcho(t *testing.T, db *gorm.DB) *echo.Echo {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	sessions := pgauth.NewRepository(db)
	check := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now}
	adminOnly := func(next echo.HandlerFunc) echo.HandlerFunc {
		return common.RequireSameOrigin()(common.RequireAdminSession(check)(next))
	}

	e := common.NewEcho()
	api := e.Group("/api")
	(&httpauth.Handler{
		LoginUC: &authcmd.LoginUsecaseImpl{
			Admin: authcmd.AdminAccount{ID: "admin", PasswordHash: string(hash)}, Verifier: infraauth.NewBcryptVerifier(),
			Sessions: sessions, Now: time.Now,
		},
		LogoutUC: &authcmd.LogoutUsecaseImpl{Sessions: sessions}, CheckUC: check, SameOrigin: common.RequireSameOrigin(),
	}).Register(api.Group("/auth"))
	repo := pgtag.NewRepository(db)
	(&httptag.Handler{
		RegisterUC: &tagcmd.RegisterUsecaseImpl{Tags: repo}, DeleteUC: &tagcmd.DeleteUsecaseImpl{Tags: repo}, AdminOnly: adminOnly,
	}).Register(api.Group("/tags"))
	return e
}

func do(e *echo.Echo, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://localhost:3000"+path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Origin", "http://localhost:3000")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func cookieOf(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == common.SessionCookieName {
			return ck
		}
	}
	t.Fatalf("no session cookie (status=%d body=%s)", rec.Code, rec.Body.String())
	return nil
}

func TestIntegration_LoginWriteLogout(t *testing.T) {
	db := connectTestDB(t)
	e := newIntegrationEcho(t, db)

	rec := do(e, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"pw"}`, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("login status = %d (body=%s)", rec.Code, rec.Body.String())
	}
	session := cookieOf(t, rec)
	t.Cleanup(func() { db.Exec("DELETE FROM admin_session WHERE id = ?", session.Value) })

	name := fmt.Sprintf("pt-auth-%06d", time.Now().UnixNano()%1_000_000)
	rec = do(e, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, name), session)
	if rec.Code != http.StatusOK {
		t.Fatalf("write with session: status = %d (body=%s)", rec.Code, rec.Body.String())
	}
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name = ?", name) })

	if rec = do(e, http.MethodPost, "/api/auth/logout", "", session); rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", rec.Code)
	}
	if rec = do(e, http.MethodPost, "/api/tags", `{"name":"after-logout"}`, session); rec.Code != http.StatusUnauthorized {
		t.Fatalf("write after logout: status = %d, want 401", rec.Code)
	}
}

func TestIntegration_Login_401(t *testing.T) {
	e := newIntegrationEcho(t, connectTestDB(t))
	if rec := do(e, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"wrong"}`, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestIntegration_Login_400(t *testing.T) {
	e := newIntegrationEcho(t, connectTestDB(t))
	if rec := do(e, http.MethodPost, "/api/auth/login", `{}`, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestIntegration_Login_429(t *testing.T) {
	e := newIntegrationEcho(t, connectTestDB(t))
	for i := 0; i < 5; i++ {
		do(e, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"wrong"}`, nil)
	}
	if rec := do(e, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"pw"}`, nil); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
}

func TestIntegration_Session_204And401(t *testing.T) {
	db := connectTestDB(t)
	e := newIntegrationEcho(t, db)
	rec := do(e, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"pw"}`, nil)
	session := cookieOf(t, rec)
	t.Cleanup(func() { db.Exec("DELETE FROM admin_session WHERE id = ?", session.Value) })

	if rec = do(e, http.MethodGet, "/api/auth/session", "", session); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec = do(e, http.MethodGet, "/api/auth/session", "", &http.Cookie{Name: common.SessionCookieName, Value: "nope"}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestIntegration_Logout_403WithoutOrigin(t *testing.T) {
	e := newIntegrationEcho(t, connectTestDB(t))
	req := httptest.NewRequest(http.MethodPost, "http://localhost:3000/api/auth/logout", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
