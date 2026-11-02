package main

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
	"time"

	"github.com/labstack/echo/v4"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	"github.com/mrstsgk/book-management-system/backend/config"
	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// run・buildEcho は非公開なので、このテストは package main に置く。

func TestRun_FailsFastOnInvalidConfig(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "verbose")

	if err := run(); err == nil {
		t.Fatal("expected an error for an invalid LOG_LEVEL")
	}
}

func TestRun_FailsWhenDBIsUnreachable(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "1")

	if err := run(); err == nil {
		t.Fatal("expected an error when the DB is unreachable")
	}
}

// buildEcho に Fake を渡して、全ルートが Handler → UseCase を通って応答すること（依存の nil・登録漏れ）と、
// 認証・入力検証の middleware が spec から掛かっていることを確かめる。DB と外部カタログには繋がない。

type stubBooks struct{}

func (stubBooks) FindByID(context.Context, bookdomain.ID) (*bookdomain.Book, error) {
	isbn, _ := bookdomain.NewISBN("9784873118703")
	bib, _ := bookdomain.NewBibliography("書名", "著者", "出版社", "202601")
	summary, _ := bookdomain.NewSummary("まとめ")
	comment, _ := bookdomain.NewComment("感想")
	rating, _ := bookdomain.NewRating(4)
	tags, _ := bookdomain.NewTagSelection(nil)
	b := bookdomain.New(isbn, bib, nil, summary, comment, rating, tags)
	b.ID, b.Version = 1, 1
	return b, nil
}
func (stubBooks) Create(_ context.Context, b *bookdomain.Book) error {
	b.ID, b.Version = 1, 1
	return nil
}
func (stubBooks) CreateAll(context.Context, []*bookdomain.Book) error { return nil }
func (stubBooks) Update(context.Context, *bookdomain.Book) error      { return nil }
func (stubBooks) Delete(context.Context, bookdomain.ID) error         { return nil }

type stubBookQuery struct{}

func (stubBookQuery) FindDetailByID(context.Context, bookdomain.ID) (*bookdomain.BookDetail, error) {
	return &bookdomain.BookDetail{ID: 1, ISBN: "9784873118703", Title: "書名", Authors: "著者", Publisher: "出版社", PublishedOn: "202601",
		Summary: "まとめ", Tags: []string{}, Comment: "感想", Rating: 4, Version: 1}, nil
}
func (stubBookQuery) Search(context.Context, bookdomain.ListCondition, bookdomain.ListRange) (*bookdomain.BookList, error) {
	return &bookdomain.BookList{Items: []*bookdomain.BookListItem{}}, nil
}
func (stubBookQuery) ExistsByISBN(context.Context, bookdomain.ISBN) (bool, error) { return false, nil }

type stubCatalog struct{}

func (stubCatalog) Lookup(_ context.Context, isbn bookdomain.ISBN) (*bookdomain.CatalogEntry, error) {
	bib, _ := bookdomain.NewBibliography("書名", "著者", "出版社", "202601")
	return &bookdomain.CatalogEntry{ISBN: isbn, Bibliography: bib}, nil
}

type stubTags struct{}

func (stubTags) FindByID(context.Context, tagdomain.ID) (*tagdomain.Tag, error) {
	name, _ := tagdomain.NewName("タグ")
	t := tagdomain.New(name)
	t.ID, t.Version = 1, 1
	return t, nil
}
func (stubTags) Create(_ context.Context, t *tagdomain.Tag) error { t.ID, t.Version = 1, 1; return nil }
func (stubTags) Update(context.Context, *tagdomain.Tag) error     { return nil }
func (stubTags) Delete(context.Context, tagdomain.ID) error       { return nil }

type stubTagQuery struct{}

func (stubTagQuery) FindAll(context.Context) (*tagdomain.TagList, error) {
	return &tagdomain.TagList{Items: []*tagdomain.TagListItem{}}, nil
}
func (stubTagQuery) ExistsAll(context.Context, []tagdomain.ID) (bool, error) { return true, nil }
func (stubTagQuery) ExistsByName(context.Context, string, tagdomain.ID) (bool, error) {
	return false, nil
}
func (stubTagQuery) CountBooks(context.Context) (*tagdomain.TagBookCounts, error) {
	return &tagdomain.TagBookCounts{Items: []*tagdomain.TagBookCount{}}, nil
}

// stubSessions は "ok" のセッションだけを有効とみなす。
type stubSessions struct{}

func (stubSessions) Create(context.Context, *authdomain.Session) error { return nil }
func (stubSessions) FindByID(_ context.Context, id authdomain.SessionID) (*authdomain.Session, error) {
	if id != "ok" {
		return nil, domainerr.ErrNotFound
	}
	return authdomain.NewSession(id, time.Now()), nil
}
func (stubSessions) Delete(context.Context, authdomain.SessionID) error { return nil }
func (stubSessions) UpdateExpiry(context.Context, authdomain.SessionID, time.Time) error {
	return nil
}

type stubVerifier struct{}

func (stubVerifier) Matches(_, password string) bool { return password == "pw" }

func newTestEcho(t *testing.T) *echo.Echo {
	t.Helper()
	spec, err := adminapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	e, err := buildEcho(spec, deps{
		Books: stubBooks{}, BookQuery: stubBookQuery{}, Catalog: stubCatalog{},
		Tags: stubTags{}, TagQuery: stubTagQuery{}, Sessions: stubSessions{},
		Admin: authcmd.AdminAccount{ID: "admin", PasswordHash: "hash"}, Verifier: stubVerifier{}, Now: time.Now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func send(e *echo.Echo, method, path, body, cookie, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	req.Header.Set("Origin", origin)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestBuildEcho_すべてのルートがHandlerとUseCaseを通って応答する(t *testing.T) {
	silenceLog(t)
	e := newTestEcho(t)
	const self = "http://example.com"
	tests := []struct {
		method, path, body, cookie string
		want                       int
	}{
		{http.MethodGet, "/api/admin/books", "", "ok", http.StatusOK},
		{http.MethodGet, "/api/admin/books/1", "", "ok", http.StatusOK},
		{http.MethodGet, "/api/admin/tags", "", "ok", http.StatusOK},
		{http.MethodGet, "/api/admin/tags/counts", "", "ok", http.StatusOK},
		{http.MethodPost, "/api/admin/books", `{"isbn":"9784873118703","summary":"まとめ","comment":"感想","rating":4}`, "ok", http.StatusOK},
		{http.MethodPut, "/api/admin/books/1", `{"summary":"まとめ","comment":"感想","rating":4,"version":1}`, "ok", http.StatusOK},
		{http.MethodDelete, "/api/admin/books/1", "", "ok", http.StatusNoContent},
		{http.MethodGet, "/api/admin/catalog/9784873118703", "", "ok", http.StatusOK},
		{http.MethodPost, "/api/admin/tags", `{"name":"タグ"}`, "ok", http.StatusOK},
		{http.MethodPut, "/api/admin/tags/1", `{"name":"タグ","version":1}`, "ok", http.StatusOK},
		{http.MethodDelete, "/api/admin/tags/1", "", "ok", http.StatusNoContent},
		{http.MethodPost, "/api/admin/auth/login", `{"id":"admin","password":"pw"}`, "", http.StatusNoContent},
		{http.MethodPost, "/api/admin/auth/logout", "", "ok", http.StatusNoContent},
		{http.MethodGet, "/api/admin/auth/session", "", "ok", http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := send(e, tt.method, tt.path, tt.body, tt.cookie, self)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body=%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestBuildEcho_管理のAPIはセッションとOriginを確かめる(t *testing.T) {
	silenceLog(t)
	e := newTestEcho(t)
	tests := []struct {
		name, method, path, body, cookie, origin string
		want                                     int
	}{
		{"書き込みはセッションが無ければ 401", http.MethodPost, "/api/admin/tags", `{"name":"タグ"}`, "", "http://example.com", http.StatusUnauthorized},
		{"書き込みは Origin が違えば 403", http.MethodPost, "/api/admin/tags", `{"name":"タグ"}`, "ok", "http://evil.example", http.StatusForbidden},
		{"カタログの確認はセッションが無ければ 401", http.MethodGet, "/api/admin/catalog/9784873118703", "", "", "http://evil.example", http.StatusUnauthorized},
		{"ログインはセッション無しで使えるが Origin は確かめる", http.MethodPost, "/api/admin/auth/login", `{"id":"admin","password":"pw"}`, "", "http://evil.example", http.StatusForbidden},
		{"参照もセッションが無ければ 401", http.MethodGet, "/api/admin/books", "", "", "http://example.com", http.StatusUnauthorized},
		{"参照は Origin を確かめない（GET は状態を変えないため）", http.MethodGet, "/api/admin/tags", "", "ok", "http://evil.example", http.StatusOK},
		{"未ログインの不正な入力には入力の形を返さず 401", http.MethodPost, "/api/admin/tags", `{"name":""}`, "", "http://example.com", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := send(e, tt.method, tt.path, tt.body, tt.cookie, tt.origin)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body=%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestBuildEcho_specに違反するリクエストは入力検証で400になる(t *testing.T) {
	silenceLog(t)
	rec := send(newTestEcho(t), http.MethodGet, "/api/admin/books?limit=0", "", "ok", "http://example.com")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
	var got httpserver.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if want := []httpserver.FieldError{{Field: "limit", Rule: "minimum"}}; !reflect.DeepEqual(got.Errors, want) {
		t.Fatalf("errors = %+v, want %+v", got.Errors, want)
	}
}

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestNewCatalog_ReturnsTheOpenBDCatalog(t *testing.T) {
	if newCatalog(config.CatalogConfig{OpenBDBaseURL: "http://openbd.test"}) == nil {
		t.Fatal("newCatalog returned nil")
	}
}
