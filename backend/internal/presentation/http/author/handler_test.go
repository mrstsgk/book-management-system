package author_test

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

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpauthor "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/author"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	authorcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
)

type fakeCreate func(ctx context.Context, cmd authorcmd.CreateCommand) (*domainauthor.Author, error)

func (f fakeCreate) Execute(ctx context.Context, cmd authorcmd.CreateCommand) (*domainauthor.Author, error) {
	return f(ctx, cmd)
}

type fakeUpdate func(ctx context.Context, cmd authorcmd.UpdateCommand) (*domainauthor.Author, error)

func (f fakeUpdate) Execute(ctx context.Context, cmd authorcmd.UpdateCommand) (*domainauthor.Author, error) {
	return f(ctx, cmd)
}

type fakeListBooks func(ctx context.Context, authorID int64) ([]*domainbook.BookSummary, error)

func (f fakeListBooks) Execute(ctx context.Context, authorID int64) ([]*domainbook.BookSummary, error) {
	return f(ctx, authorID)
}

func newAuthor(t *testing.T, id int64, name string, birthDate *time.Time, version int) *domainauthor.Author {
	t.Helper()
	n, err := domainauthor.NewName(name)
	if err != nil {
		t.Fatal(err)
	}
	a := &domainauthor.Author{ID: domainauthor.ID(id), Name: n, Version: version}
	if birthDate != nil {
		b := domainauthor.RestoreBirthDate(*birthDate)
		a.BirthDate = &b
	}
	return a
}

// serve wires the handler the same way cmd/api/main.go does (NewEcho + Register).
func serve(t *testing.T, h *httpauthor.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	e := common.NewEcho()
	h.Register(e.Group("/api/authors"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return v
}

func ptr[T any](v T) *T { return &v }

func mustNotCallCreate(t *testing.T) fakeCreate {
	return func(context.Context, authorcmd.CreateCommand) (*domainauthor.Author, error) {
		t.Error("usecase must not be called")
		return nil, nil
	}
}

func TestHandlerCreate(t *testing.T) {
	birth := time.Date(1909, 6, 19, 0, 0, 0, 0, time.UTC)

	t.Run("生年月日付きで作成し200で著者を返す", func(t *testing.T) {
		var got authorcmd.CreateCommand
		h := &httpauthor.Handler{CreateUC: fakeCreate(func(_ context.Context, cmd authorcmd.CreateCommand) (*domainauthor.Author, error) {
			got = cmd
			return newAuthor(t, 1, cmd.Name, cmd.BirthDate, 1), nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/authors", `{"name":"太宰治","birthDate":"1909-06-19"}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.Name != "太宰治" || got.BirthDate == nil || !got.BirthDate.Equal(birth) {
			t.Fatalf("usecase received %+v", got)
		}
		want := httpauthor.Response{ID: 1, Name: "太宰治", BirthDate: ptr("1909-06-19"), Version: 1}
		if res := decode[httpauthor.Response](t, rec); !reflect.DeepEqual(res, want) {
			t.Fatalf("body = %+v, want %+v", res, want)
		}
	})

	t.Run("生年月日なしはnullで返す", func(t *testing.T) {
		h := &httpauthor.Handler{CreateUC: fakeCreate(func(_ context.Context, cmd authorcmd.CreateCommand) (*domainauthor.Author, error) {
			if cmd.BirthDate != nil {
				t.Errorf("BirthDate = %v, want nil", cmd.BirthDate)
			}
			return newAuthor(t, 2, cmd.Name, nil, 1), nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/authors", `{"name":"芥川龍之介"}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"birthDate":null`) {
			t.Fatalf("body = %s, want birthDate null", rec.Body.String())
		}
	})

	invalid := []struct {
		name string
		body string
		want []common.FieldError
	}{
		{name: "名前なしは400", body: `{}`, want: []common.FieldError{{Field: "name", Rule: "required"}}},
		{name: "名前101文字は400", body: `{"name":"` + strings.Repeat("あ", 101) + `"}`, want: []common.FieldError{{Field: "name", Rule: "max"}}},
		{name: "日付形式でない生年月日は400", body: `{"name":"太宰治","birthDate":"1909/06/19"}`, want: []common.FieldError{{Field: "birthDate", Rule: "datetime"}}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, &httpauthor.Handler{CreateUC: mustNotCallCreate(t)}, http.MethodPost, "/api/authors", tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := decode[common.ErrorResponse](t, rec).Errors; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("ドメインルール違反は400", func(t *testing.T) {
		h := &httpauthor.Handler{CreateUC: fakeCreate(func(context.Context, authorcmd.CreateCommand) (*domainauthor.Author, error) {
			return nil, domaincommon.ErrInvalid
		})}
		if rec := serve(t, h, http.MethodPost, "/api/authors", `{"name":"太宰治","birthDate":"2999-01-01"}`); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestHandlerUpdate(t *testing.T) {
	t.Run("パスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got authorcmd.UpdateCommand
		h := &httpauthor.Handler{UpdateUC: fakeUpdate(func(_ context.Context, cmd authorcmd.UpdateCommand) (*domainauthor.Author, error) {
			got = cmd
			return newAuthor(t, cmd.ID, cmd.Name, nil, cmd.Version+1), nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/authors/7", `{"name":"太宰治","version":3}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.ID != 7 || got.Version != 3 || got.Name != "太宰治" {
			t.Fatalf("usecase received %+v", got)
		}
		if res := decode[httpauthor.Response](t, rec); res.ID != 7 || res.Version != 4 {
			t.Fatalf("body = %+v", res)
		}
	})

	t.Run("バージョンなしは400", func(t *testing.T) {
		h := &httpauthor.Handler{UpdateUC: fakeUpdate(func(context.Context, authorcmd.UpdateCommand) (*domainauthor.Author, error) {
			t.Error("usecase must not be called")
			return nil, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/authors/7", `{"name":"太宰治"}`)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if want := []common.FieldError{{Field: "version", Rule: "required"}}; !reflect.DeepEqual(decode[common.ErrorResponse](t, rec).Errors, want) {
			t.Fatalf("body = %s", rec.Body.String())
		}
	})

	statuses := []struct {
		name string
		path string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/authors/abc", want: http.StatusBadRequest},
		{name: "存在しない著者は404", path: "/api/authors/7", err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/authors/7", err: domaincommon.ErrConflict, want: http.StatusConflict},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpauthor.Handler{UpdateUC: fakeUpdate(func(context.Context, authorcmd.UpdateCommand) (*domainauthor.Author, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodPut, tt.path, `{"name":"太宰治","version":1}`); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerListBooks(t *testing.T) {
	t.Run("著者の書籍一覧を配列で返す", func(t *testing.T) {
		var gotID int64
		h := &httpauthor.Handler{BooksUC: fakeListBooks(func(_ context.Context, id int64) ([]*domainbook.BookSummary, error) {
			gotID = id
			return []*domainbook.BookSummary{{ID: 1, Title: "人間失格", Price: 1500, Status: 2}}, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/authors/3/books", "")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if gotID != 3 {
			t.Fatalf("usecase received id=%d, want 3", gotID)
		}
		want := []httpauthor.BookResponse{{ID: 1, Title: "人間失格", Price: 1500, Status: 2}}
		if got := decode[[]httpauthor.BookResponse](t, rec); !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v, want %+v", got, want)
		}
	})

	t.Run("書籍が無ければ空配列", func(t *testing.T) {
		h := &httpauthor.Handler{BooksUC: fakeListBooks(func(context.Context, int64) ([]*domainbook.BookSummary, error) {
			return nil, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/authors/3/books", "")

		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Fatalf("status = %d body = %s, want 200 []", rec.Code, rec.Body.String())
		}
	})

	t.Run("存在しない著者は404", func(t *testing.T) {
		h := &httpauthor.Handler{BooksUC: fakeListBooks(func(context.Context, int64) ([]*domainbook.BookSummary, error) {
			return nil, domaincommon.ErrNotFound
		})}
		if rec := serve(t, h, http.MethodGet, "/api/authors/3/books", ""); rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}
