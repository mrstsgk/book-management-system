package admin_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	"github.com/mrstsgk/book-management-system/backend/internal/apitest"
	authbcrypt "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/bcrypt"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/postgres"
	authadmin "github.com/mrstsgk/book-management-system/backend/internal/auth/presentation/http/admin"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagadmin "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/admin"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
)

func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgplatform.Connect(pgplatform.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up migrate-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM admin_session LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// testServer は、ログインの Handler と、セッションが要る書き込みの例として分野タグの管理の Handler を載せる。
type testServer struct {
	apitest.AdminUnimplemented
	*authadmin.Handler
	tagHandler
}

type tagHandler = *tagadmin.Handler

// newServer は、実 DB のセッションと bcrypt を配線する。セッションの確認も本物を使う（ログイン → 書き込み → ログアウトの配線を確かめるため）。
func newServer(t *testing.T, db *gorm.DB) *apitest.Server {
	t.Helper()
	// テストが速いよう、bcrypt は最小のコストで作る
	hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	sessions := pgauth.NewRepository(db)
	repo := pgtag.NewRepository(db)
	return apitest.NewAdminServer(t, &testServer{
		Handler: &authadmin.Handler{
			LoginUsecase: &authcmd.LoginUsecaseImpl{
				Admin: authcmd.AdminAccount{ID: "admin", PasswordHash: string(hash)}, Verifier: authbcrypt.NewBcryptVerifier(),
				Sessions: sessions, Now: time.Now,
			},
			LogoutUsecase: &authcmd.LogoutUsecaseImpl{Sessions: sessions},
		},
		tagHandler: &tagadmin.Handler{
			RegisterTagUsecase: &tagcmd.RegisterTagUsecaseImpl{Tags: repo, TagQuery: pgtag.NewQuery(db)},
		},
	}, &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now})
}

func sessionOf(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == httpserver.SessionCookieName && ck.Value != "" {
			return ck.Value
		}
	}
	t.Fatalf("no session cookie (status=%d body=%s)", rec.Code, rec.Body.String())
	return ""
}

func login(t *testing.T, db *gorm.DB, s *apitest.Server) string {
	t.Helper()
	rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":"admin","password":"pw"}`})
	apitest.AssertStatus(t, rec, http.StatusNoContent)
	session := sessionOf(t, rec)
	t.Cleanup(func() { db.Exec("DELETE FROM admin_session WHERE id = ?", session) })
	return session
}

func TestLogin(t *testing.T) {
	db := connectTestDB(t)

	t.Run("204: ログインで発行した Cookie で書き込めて、ログアウト後は 401", func(t *testing.T) {
		s := newServer(t, db)
		session := login(t, db, s)

		name := fmt.Sprintf("pt-auth-%09d", time.Now().UnixNano()%1_000_000_000)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: session, Body: fmt.Sprintf(`{"name":%q}`, name)})
		apitest.AssertStatus(t, rec, http.StatusOK)
		t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name = ?", name) })

		apitest.AssertStatus(t, s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/logout", Cookie: session}), http.StatusNoContent)
		rec = s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: session, Body: `{"name":"after-logout"}`})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("400: ID が 0 文字なら minLength、パスワードが無ければ required", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":""}`})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("id", "minLength"), apitest.FieldError("password", "required"))
	})

	t.Run("通す: ID とパスワードが 1 文字（違えば 401）", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":"a","password":"b"}`})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "loginFailed")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Origin: "http://evil.example", Body: `{"id":"admin","password":"pw"}`})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("429: 5 回続けて失敗すると、正しくても tooManyLoginAttempts と待つ秒数を返す", func(t *testing.T) {
		s := newServer(t, db)
		for range 5 {
			s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":"admin","password":"wrong"}`})
		}
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":"admin","password":"pw"}`})
		apitest.AssertCode(t, rec, http.StatusTooManyRequests, "tooManyLoginAttempts")
		body := apitest.Decode[adminapi.ErrorResponse](t, rec)
		if body.Params == nil {
			t.Fatal("params must carry retryAfterSeconds")
		}
		if secs, ok := (*body.Params)["retryAfterSeconds"].(float64); !ok || secs < 1 || secs > 60 {
			t.Fatalf("retryAfterSeconds = %v, want 1..60", (*body.Params)["retryAfterSeconds"])
		}
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/login", Body: `{"id":"admin","password":"pw"}`, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestLogout(t *testing.T) {
	db := connectTestDB(t)

	t.Run("204: ログインしていなくても成功する", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/logout"})
		apitest.AssertStatus(t, rec, http.StatusNoContent)
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/logout", Origin: "http://evil.example"})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/auth/logout", Cookie: "some-session", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestCheckSession(t *testing.T) {
	db := connectTestDB(t)

	t.Run("204: 有効なセッション", func(t *testing.T) {
		s := newServer(t, db)
		session := login(t, db, s)
		apitest.AssertStatus(t, s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/auth/session", Cookie: session}), http.StatusNoContent)
	})

	t.Run("401: 無効なセッションは unauthenticated", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/auth/session", Cookie: "nope"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/auth/session", Cookie: "nope", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}
