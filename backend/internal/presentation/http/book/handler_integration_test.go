package book_test

// Presentation → Infrastructure の結合テスト（docs/rules/testing.md）。
// Handler → UseCase → Repository/Query → 実 PostgreSQL の配線が噛み合っていることを、
// 各エンドポイントが返しうる HTTP ステータスごとに確認する（内部の分岐網羅は目的にしない。
// 分岐網羅・境界値は既存の Fake ベースの handler_test.go・usecase の Fake テスト・
// postgres/book の契約テストが担う）。外部カタログ（openBD・楽天）だけは Fake に差し替える
// （gateway パッケージで別途検証済みのため、ここで実ネットワークに頼る必要はない）。

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// connectIntegrationDB はローカルの開発用 DB に接続する。実際の PostgreSQL に対する結合テストなので、
// DB に繋がらなければ失敗ではなく skip する（契約テストと同じ理由。docs/rules/testing.md）。
func connectIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgcommon.Connect(pgcommon.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up migrate-up` first): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM book LIMIT 0").Error; err != nil {
		t.Skipf("skipping: schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// integrationFakeCatalog は外部カタログ（openBD・楽天）を差し替える手書き Fake。
// gateway/openbd・gateway/rakuten 自体は別途検証済みのため、ここでは実ネットワークに頼らない。
type integrationFakeCatalog struct {
	entry *domainbook.CatalogEntry
	err   error
}

func (f *integrationFakeCatalog) Lookup(context.Context, domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	return f.entry, f.err
}

// newIntegrationHandler は本物の Repository/Query/UseCase を実DBに配線した Handler を返す（AdminOnly は serve() が付ける）。
func newIntegrationHandler(db *gorm.DB, catalog domainbook.BookCatalog) *httpbook.Handler {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db, time.Now)
	tagQuery := pgtag.NewQuery(db)
	return &httpbook.Handler{
		RegisterUC:       &bookcmd.RegisterUsecaseImpl{Books: books, Catalog: catalog, Details: bookQuery, Tags: tagQuery},
		UpdateUC:         &bookcmd.UpdateUsecaseImpl{Books: books, Catalog: catalog, Details: bookQuery, Tags: tagQuery, Now: time.Now},
		DeleteUC:         &bookcmd.DeleteUsecaseImpl{Books: books},
		DisableRakutenUC: &bookcmd.DisableRakutenUsecaseImpl{Books: books},
		GetUC:            &bookqry.GetUsecaseImpl{Books: bookQuery},
		ListUC:           &bookqry.ListUsecaseImpl{Books: bookQuery},
	}
}

// integrationCatalogEntry は Fake カタログが返す書誌・書影を組み立てる。
func integrationCatalogEntry(t *testing.T, isbn string) *domainbook.CatalogEntry {
	t.Helper()
	parsed, err := domainbook.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := domainbook.NewBibliography("結合テストの書名", "結合テスト著者", "結合テスト出版社", "202609")
	if err != nil {
		t.Fatal(err)
	}
	return &domainbook.CatalogEntry{ISBN: parsed, Bibliography: bib, Cover: nil}
}

// mustCreateIntegrationBook は Repository へ直接（外部カタログを経由せず）本を1冊作る。
// 事前データの用意が目的で、登録APIそのもののテストではないため。
func mustCreateIntegrationBook(t *testing.T, db *gorm.DB, isbn string) *domainbook.Book {
	t.Helper()
	entry := integrationCatalogEntry(t, isbn)
	summary, err := domainbook.NewSummary("結合テスト用の一言まとめ")
	if err != nil {
		t.Fatal(err)
	}
	comment, err := domainbook.NewComment("結合テスト用の感想")
	if err != nil {
		t.Fatal(err)
	}
	rating, err := domainbook.NewRating(4)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := domainbook.NewTagSelection(nil)
	if err != nil {
		t.Fatal(err)
	}
	b := domainbook.New(entry.ISBN, entry.Bibliography, entry.Cover, summary, comment, rating, tags)
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("create book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
	return b
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

func TestBookHandlerIntegration_List(t *testing.T) {
	db := connectIntegrationDB(t)
	mustCreateIntegrationBook(t, db, "9780000004000")
	h := newIntegrationHandler(db, &integrationFakeCatalog{})

	t.Run("200: ハッピーパス", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[httpbook.ListResponse](t, rec)
		if got.Total < 1 || len(got.Items) < 1 {
			t.Fatalf("body = %+v, want at least 1 item", got)
		}
	})

	t.Run("400: qが不正", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books?tagId=0", "", false)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_Get(t *testing.T) {
	db := connectIntegrationDB(t)
	b := mustCreateIntegrationBook(t, db, "9780000004017")
	h := newIntegrationHandler(db, &integrationFakeCatalog{})

	t.Run("200: ハッピーパス", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books/"+itoa(int64(b.ID)), "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[httpbook.Response](t, rec)
		if got.ISBN != "9780000004017" {
			t.Fatalf("isbn = %q, want 9780000004017", got.ISBN)
		}
	})

	t.Run("404: 存在しないID", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books/999999999", "", false)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_Register(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		const isbn = "9780000004024"
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		body := `{"isbn":"` + isbn + `","summary":"要約","comment":"良書","rating":5}`
		rec := serve(t, h, http.MethodPost, "/api/books", body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[httpbook.Response](t, rec)
		t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", got.ID) })
		if got.ISBN != isbn {
			t.Fatalf("isbn = %q, want %s", got.ISBN, isbn)
		}
	})

	t.Run("400: バリデーション不正", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"123","summary":"要約","comment":"良書","rating":5}`, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("401: トークン無し", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"9780000004031","summary":"要約","comment":"良書","rating":5}`, false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("409: 同じISBNの登録済み", func(t *testing.T) {
		const isbn = "9780000004048"
		mustCreateIntegrationBook(t, db, isbn)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"`+isbn+`","summary":"要約","comment":"良書","rating":5}`, true)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_Update(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, "9780000004055")
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, "9780000004055")})
		body := `{"summary":"読み返した要約","comment":"読み返した感想","rating":3,"version":` + itoa(int64(b.Version)) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[httpbook.Response](t, rec)
		if got.Comment != "読み返した感想" {
			t.Fatalf("comment = %q, want 読み返した感想", got.Comment)
		}
	})

	t.Run("400: バリデーション不正", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, "9780000004062")
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, "9780000004062")})
		body := `{"summary":"要約","comment":"","rating":3,"version":` + itoa(int64(b.Version)) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("401: トークン無し", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodPut, "/api/books/1", `{"summary":"要約","comment":"良書","rating":3,"version":1}`, false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("404: 存在しないID", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodPut, "/api/books/999999999", `{"summary":"要約","comment":"良書","rating":3,"version":1}`, true)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("409: バージョン不一致", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, "9780000004079")
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, "9780000004079")})
		body := `{"summary":"要約","comment":"良書","rating":3,"version":` + itoa(int64(b.Version)+1) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_Delete(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("204: ハッピーパス", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, "9780000004086")
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/"+itoa(int64(b.ID)), "", true)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("401: トークン無し", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/1", "", false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("404: 存在しないID", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/999999999", "", true)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_DisableRakuten(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("204: ハッピーパス", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, "9780000004093")
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/"+itoa(int64(b.ID))+"/rakuten", "", true)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("401: トークン無し", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/1/rakuten", "", false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("404: 存在しないID", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/999999999/rakuten", "", true)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}
