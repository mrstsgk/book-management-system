package admin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	bookadmin "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/admin"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
)

// connectIntegrationDB はローカルの開発用 DB に接続する。実際の PostgreSQL に対する結合テストなので、
// DB に繋がらなければ skip せず失敗にする（契約テストと同じ理由。docs/rules/testing.md）。
func connectIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgplatform.Connect(pgplatform.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up migrate-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM book LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// integrationFakeCatalog は外部カタログ（openBD）を差し替える手書き Fake。
// gateway/openbd 自体は別途検証済みのため、ここでは実ネットワークに頼らない。
type integrationFakeCatalog struct {
	entry *bookdomain.CatalogEntry
	err   error
}

func (f *integrationFakeCatalog) Lookup(context.Context, bookdomain.ISBN) (*bookdomain.CatalogEntry, error) {
	return f.entry, f.err
}

// newIntegrationHandler は本物の Repository/Query/UseCase を実DBに配線した Handler を返す（AdminOnly は serve() が付ける）。
func newIntegrationHandler(db *gorm.DB, catalog bookdomain.BookCatalog) *bookadmin.Handler {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)
	tagQuery := pgtag.NewQuery(db)
	return &bookadmin.Handler{
		RegisterBookUsecase: &bookcmd.RegisterBookUsecaseImpl{Books: books, Catalog: catalog, Details: bookQuery, Tags: tagQuery},
		UpdateBookUsecase:   &bookcmd.UpdateBookUsecaseImpl{Books: books, Catalog: catalog, Details: bookQuery, Tags: tagQuery},
		DeleteBookUsecase:   &bookcmd.DeleteBookUsecaseImpl{Books: books},
	}
}

// integrationCatalogEntry は Fake カタログが返す書誌・書影を組み立てる。
func integrationCatalogEntry(t *testing.T, isbn string) *bookdomain.CatalogEntry {
	t.Helper()
	parsed, err := bookdomain.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := bookdomain.NewBibliography("結合テストの書名", "結合テスト著者", "結合テスト出版社", "202609")
	if err != nil {
		t.Fatal(err)
	}
	return &bookdomain.CatalogEntry{ISBN: parsed, Bibliography: bib, Cover: nil}
}

// mustCreateIntegrationBook は Repository へ直接（外部カタログを経由せず）本を1冊作る。
// 事前データの用意が目的で、登録APIそのもののテストではないため。
func mustCreateIntegrationBook(t *testing.T, db *gorm.DB, isbn string) *bookdomain.Book {
	t.Helper()
	entry := integrationCatalogEntry(t, isbn)
	summary, err := bookdomain.NewSummary("結合テスト用の一言まとめ")
	if err != nil {
		t.Fatal(err)
	}
	comment, err := bookdomain.NewComment("結合テスト用の感想")
	if err != nil {
		t.Fatal(err)
	}
	rating, err := bookdomain.NewRating(4)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := bookdomain.NewTagSelection(nil)
	if err != nil {
		t.Fatal(err)
	}
	b := bookdomain.New(entry.ISBN, entry.Bibliography, entry.Cover, summary, comment, rating, tags)
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("create book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
	return b
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

// integrationISBNSeq は nextIntegrationISBN の一意性を同一プロセス内の複数呼び出しでも保証するための連番。
var integrationISBNSeq atomic.Int64

// nextIntegrationISBN は結合テストごとに一意な、チェックディジットが正しい ISBN-13 を生成する。
// 固定リテラルだと、前回実行のクリーンアップ漏れ（プロセスの強制終了など）で同じ ISBN の行が
// 開発用 DB に残っていた場合、Repository がユニーク制約違反を ErrConflict に変換してしまい、
// Handler の検証（本テストの目的）に届く前にテストが落ちてしまうため。
func nextIntegrationISBN(t *testing.T) string {
	t.Helper()
	n := time.Now().UnixNano() + integrationISBNSeq.Add(1)
	first12 := fmt.Sprintf("978%09d", n%1_000_000_000)
	sum := 0
	for i, r := range first12 {
		d := int(r - '0')
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return first12 + strconv.Itoa((10-sum%10)%10)
}

func TestBookHandlerIntegration_Register(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		isbn := nextIntegrationISBN(t)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		body := `{"isbn":"` + isbn + `","summary":"要約","comment":"良書","rating":5}`
		rec := serve(t, h, http.MethodPost, "/api/books", body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[bookpublic.Response](t, rec)
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

	t.Run("401: 未ログイン", func(t *testing.T) {
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"`+nextIntegrationISBN(t)+`","summary":"要約","comment":"良書","rating":5}`, false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("400: 同じISBNの登録済みは相関ルール isbnDuplicated", func(t *testing.T) {
		isbn := nextIntegrationISBN(t)
		mustCreateIntegrationBook(t, db, isbn)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"`+isbn+`","summary":"要約","comment":"良書","rating":5}`, true)
		assertFieldErrors(t, rec, []httpserver.FieldError{{Field: "isbn", Rule: "isbnDuplicated"}})
	})

	t.Run("400: 存在しない分野タグは相関ルール tagIdsNotFound", func(t *testing.T) {
		isbn := nextIntegrationISBN(t)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		rec := serve(t, h, http.MethodPost, "/api/books", `{"isbn":"`+isbn+`","summary":"要約","tagIds":[999999999],"comment":"良書","rating":5}`, true)
		assertFieldErrors(t, rec, []httpserver.FieldError{{Field: "tagIds", Rule: "tagIdsNotFound"}})
		var count int64
		db.Table("book").Where("isbn = ?", isbn).Count(&count)
		if count != 0 {
			t.Fatal("the book must not be saved when a violation exists")
		}
	})
}

func TestBookHandlerIntegration_Update(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		isbn := nextIntegrationISBN(t)
		b := mustCreateIntegrationBook(t, db, isbn)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		body := `{"summary":"読み返した要約","comment":"読み返した感想","rating":3,"version":` + itoa(int64(b.Version)) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[bookpublic.Response](t, rec)
		if got.Comment != "読み返した感想" {
			t.Fatalf("comment = %q, want 読み返した感想", got.Comment)
		}
	})

	t.Run("400: バリデーション不正", func(t *testing.T) {
		isbn := nextIntegrationISBN(t)
		b := mustCreateIntegrationBook(t, db, isbn)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		body := `{"summary":"要約","comment":"","rating":3,"version":` + itoa(int64(b.Version)) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
		assertIntegrationBookUnchanged(t, db, b)
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
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
		isbn := nextIntegrationISBN(t)
		b := mustCreateIntegrationBook(t, db, isbn)
		h := newIntegrationHandler(db, &integrationFakeCatalog{entry: integrationCatalogEntry(t, isbn)})
		body := `{"summary":"要約","comment":"良書","rating":3,"version":` + itoa(int64(b.Version)+1) + `}`
		rec := serve(t, h, http.MethodPut, "/api/books/"+itoa(int64(b.ID)), body, true)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
		}
		assertIntegrationBookUnchanged(t, db, b)
	})
}

// assertIntegrationBookUnchanged は、失敗したはずの更新リクエストの後で本の内容・バージョンが
// 呼び出し前のまま変わっていないことを確認する（docs/rules/testing.md の「失敗時の非変化」）。
func assertIntegrationBookUnchanged(t *testing.T, db *gorm.DB, before *bookdomain.Book) {
	t.Helper()
	after, err := pgbook.NewRepository(db).FindByID(context.Background(), before.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if after.Comment.String() != before.Comment.String() || after.Summary.String() != before.Summary.String() ||
		after.Rating.Int() != before.Rating.Int() || after.Version != before.Version {
		t.Fatalf("book changed on a failed update: got comment=%q summary=%q rating=%d version=%d, want comment=%q summary=%q rating=%d version=%d",
			after.Comment.String(), after.Summary.String(), after.Rating.Int(), after.Version,
			before.Comment.String(), before.Summary.String(), before.Rating.Int(), before.Version)
	}
}

func TestBookHandlerIntegration_Delete(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("204: ハッピーパス", func(t *testing.T) {
		b := mustCreateIntegrationBook(t, db, nextIntegrationISBN(t))
		h := newIntegrationHandler(db, &integrationFakeCatalog{})
		rec := serve(t, h, http.MethodDelete, "/api/books/"+itoa(int64(b.ID)), "", true)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
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

// assertFieldErrors は 400 と errors[{field, rule}] の内容（順序込み）を確かめる。
func assertFieldErrors(t *testing.T, rec *httptest.ResponseRecorder, want []httpserver.FieldError) {
	t.Helper()
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
	var e httpserver.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	if !reflect.DeepEqual(e.Errors, want) {
		t.Fatalf("errors = %+v, want %+v", e.Errors, want)
	}
}
