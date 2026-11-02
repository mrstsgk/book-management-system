package admin_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	"github.com/mrstsgk/book-management-system/backend/internal/apitest"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	bookadmin "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/admin"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
)

// connectIntegrationDB はローカルの開発用 DB に接続する。実際の PostgreSQL に対する結合テストなので、
// DB に繋がらなければ skip せず失敗にする（docs/rules/testing.md）。
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

// fakeCatalog は外部カタログ（openBD）を差し替える手書き Fake。
// gateway/openbd 自体は別途検証済みのため、ここでは実ネットワークに頼らない。
type fakeCatalog struct {
	entry *bookdomain.CatalogEntry
	err   error
}

func (f *fakeCatalog) Lookup(context.Context, bookdomain.ISBN) (*bookdomain.CatalogEntry, error) {
	return f.entry, f.err
}

// skipDuplicateCheck は、ISBN の重複の事前検査をすり抜ける Query。同時登録で DB の一意制約（最後の砦）に当たる状況を作る。
type skipDuplicateCheck struct{ bookdomain.Query }

func (skipDuplicateCheck) ExistsByISBN(context.Context, bookdomain.ISBN) (bool, error) {
	return false, nil
}

type testServer struct {
	apitest.AdminUnimplemented
	*bookadmin.Handler
	*bookadmin.CatalogHandler
}

// newServer は、本物の Repository/Query/UseCase を実 DB に配線した管理の Handler をサーバーに載せる。
func newServer(t *testing.T, db *gorm.DB, catalog bookdomain.BookCatalog, details bookdomain.Query) *apitest.Server {
	t.Helper()
	books := pgbook.NewRepository(db)
	if details == nil {
		details = pgbook.NewQuery(db)
	}
	tagQuery := pgtag.NewQuery(db)
	return apitest.NewAdminServer(t, &testServer{
		Handler: &bookadmin.Handler{
			ReadBookUsecase:     &bookqry.ReadBookUsecaseImpl{Books: details},
			SearchBooksUsecase:  &bookqry.SearchBooksUsecaseImpl{Books: details},
			RegisterBookUsecase: &bookcmd.RegisterBookUsecaseImpl{Books: books, Catalog: catalog, Details: details, Tags: tagQuery},
			UpdateBookUsecase:   &bookcmd.UpdateBookUsecaseImpl{Books: books, Catalog: catalog, Details: details, Tags: tagQuery},
			DeleteBookUsecase:   &bookcmd.DeleteBookUsecaseImpl{Books: books},
		},
		CatalogHandler: &bookadmin.CatalogHandler{LookupCatalogUsecase: &bookqry.LookupCatalogUsecaseImpl{Catalog: catalog}},
	}, apitest.AcceptSession{})
}

// catalogEntry は Fake カタログが返す書誌（書影なし）を組み立てる。
func catalogEntry(t *testing.T, isbn string) *bookdomain.CatalogEntry {
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

// mustCreateBook は Repository へ直接（外部カタログを経由せず）本を 1 冊作る。登録 API のテストではない準備のため。
func mustCreateBook(t *testing.T, db *gorm.DB, isbn string) *bookdomain.Book {
	t.Helper()
	entry := catalogEntry(t, isbn)
	summary, _ := bookdomain.NewSummary("結合テスト用の一言まとめ")
	comment, _ := bookdomain.NewComment("結合テスト用の感想")
	rating, _ := bookdomain.NewRating(4)
	tags, _ := bookdomain.NewTagSelection(nil)
	b := bookdomain.New(entry.ISBN, entry.Bibliography, entry.Cover, summary, comment, rating, tags)
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("create book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
	return b
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

var isbnSeq atomic.Int64

// nextISBN は、テストごとに一意でチェックディジットが正しい ISBN-13 を作る。
// 固定値だと、前回の実行の片付け漏れで同じ ISBN が DB に残っていたときに、検証したいこと以外で落ちるため。
func nextISBN(t *testing.T) string {
	t.Helper()
	n := time.Now().UnixNano() + isbnSeq.Add(1)
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

func registerBody(isbn string) string {
	return `{"isbn":"` + isbn + `","summary":"要約","comment":"良書","rating":5}`
}

func TestRegisterBook(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: 登録した本を返し、DB に全項目が入る", func(t *testing.T) {
		isbn := nextISBN(t)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session,
			Body: `{"isbn":"` + isbn + `","summary":"要約","titleOverride":"上書きの書名","comment":"良書","rating":5}`})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.BookResponse](t, rec)
		t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", got.ID) })
		if got.Isbn != isbn || got.Title != "上書きの書名" || got.Summary != "要約" || got.Comment != "良書" || got.Rating != 5 ||
			got.Authors != "結合テスト著者" || got.Publisher != "結合テスト出版社" || got.PublishedOn != "202609" ||
			got.TitleOverride == nil || *got.TitleOverride != "上書きの書名" || len(got.Tags) != 0 || got.Version != 1 || got.CoverURL != nil {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: 独立した入力・相関ルールの違反をまとめて返し、保存しない", func(t *testing.T) {
		isbn := nextISBN(t)
		mustCreateBook(t, db, isbn)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session,
			Body: `{"isbn":"` + isbn + `","summary":"要約","tagIds":[999999999],"comment":"良書","rating":5}`})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("isbn", "isbnDuplicated"), apitest.FieldError("tagIds", "tagIdsNotFound"))
	})

	t.Run("400: 外部カタログに無い ISBN は isbnNotInCatalog", func(t *testing.T) {
		isbn := nextISBN(t)
		s := newServer(t, db, &fakeCatalog{err: domainerr.New(domainerr.ErrNotFound, "catalogEntryNotFound", "無い")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: registerBody(isbn)})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("isbn", "isbnNotInCatalog"))
		var count int64
		db.Table("book").Where("isbn = ?", isbn).Count(&count)
		if count != 0 {
			t.Fatal("the book must not be saved when a violation exists")
		}
	})

	t.Run("400: ISBN の形式が合わなければ format（外部カタログは照会しない）", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{err: errors.New("must not be called")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: registerBody("9784873118700")})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("isbn", "format"))
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Body: registerBody(nextISBN(t))})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Origin: "http://evil.example", Body: registerBody(nextISBN(t))})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("409: 事前の検査をすり抜けた同時登録は一意制約で isbnDuplicated", func(t *testing.T) {
		isbn := nextISBN(t)
		mustCreateBook(t, db, isbn)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, skipDuplicateCheck{pgbook.NewQuery(db)})
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: registerBody(isbn)})
		apitest.AssertCode(t, rec, http.StatusConflict, "isbnDuplicated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, "9784873118703")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: registerBody(nextISBN(t)), Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

// TestRegisterBook_Constraints は、YAML に書いた入力の制約 1 つにつき、通る値と弾かれる値を 1 組ずつ確かめる。
func TestRegisterBook_Constraints(t *testing.T) {
	db := connectIntegrationDB(t)
	s := newServer(t, db, &fakeCatalog{err: errors.New("catalog is not reached in constraint checks")}, nil)
	// 他の項目が正しい本文に、1 項目だけ差し替える
	body := func(field, value string) string {
		m := map[string]string{"isbn": `"9784873118703"`, "summary": `"要約"`, "comment": `"良書"`, "rating": `5`}
		if value == "" {
			delete(m, field)
		} else {
			m[field] = value
		}
		parts := make([]string, 0, len(m))
		for k, v := range m {
			parts = append(parts, `"`+k+`":`+v)
		}
		return "{" + strings.Join(parts, ",") + "}"
	}
	tagIDs := func(n int) string {
		ids := make([]string, n)
		for i := range ids {
			ids[i] = strconv.Itoa(900000000 + i)
		}
		return "[" + strings.Join(ids, ",") + "]"
	}
	tests := []struct {
		name, field, value, rule string
	}{
		{"isbn 欠落", "isbn", "", "required"},
		{"isbn 0 文字", "isbn", `""`, "minLength"},
		{"isbn 18 文字", "isbn", `"978-4-87311-870-33"`, "maxLength"},
		{"summary 欠落", "summary", "", "required"},
		{"summary 0 文字", "summary", `""`, "minLength"},
		{"summary 101 文字", "summary", `"` + strings.Repeat("あ", 101) + `"`, "maxLength"},
		{"titleOverride 256 文字", "titleOverride", `"` + strings.Repeat("あ", 256) + `"`, "maxLength"},
		{"tagIds 11 個", "tagIds", tagIDs(11), "maxItems"},
		{"comment 欠落", "comment", "", "required"},
		{"comment 0 文字", "comment", `""`, "minLength"},
		{"comment 5001 文字", "comment", `"` + strings.Repeat("あ", 5001) + `"`, "maxLength"},
		{"rating 欠落", "rating", "", "required"},
		{"rating 0", "rating", `0`, "minimum"},
		{"rating 6", "rating", `6`, "maximum"},
		{"rating 文字列", "rating", `"5"`, "type"},
	}
	for _, tt := range tests {
		t.Run("弾く: "+tt.name, func(t *testing.T) {
			rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: body(tt.field, tt.value)})
			apitest.AssertFieldErrors(t, rec, apitest.FieldError(tt.field, tt.rule))
		})
	}
	passes := []struct {
		name, field, value, rule string
	}{
		{"isbn 1 文字", "isbn", `"9"`, "minLength"},
		{"isbn 17 文字", "isbn", `"978-4-87311-870-3"`, "maxLength"},
		{"summary 1 文字", "summary", `"あ"`, "minLength"},
		{"summary 100 文字", "summary", `"` + strings.Repeat("あ", 100) + `"`, "maxLength"},
		{"titleOverride 255 文字", "titleOverride", `"` + strings.Repeat("あ", 255) + `"`, "maxLength"},
		{"tagIds 10 個", "tagIds", tagIDs(10), "maxItems"},
		{"comment 1 文字", "comment", `"あ"`, "minLength"},
		{"comment 5000 文字", "comment", `"` + strings.Repeat("あ", 5000) + `"`, "maxLength"},
		{"rating 1", "rating", `1`, "minimum"},
		{"rating 5", "rating", `5`, "maximum"},
	}
	for _, tt := range passes {
		t.Run("通す: "+tt.name, func(t *testing.T) {
			rec := s.Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/books", Cookie: apitest.Session, Body: body(tt.field, tt.value)})
			apitest.AssertNoFieldError(t, rec, tt.field, tt.rule)
		})
	}
}

func updateBody(version int) string {
	return `{"summary":"読み返した要約","comment":"読み返した感想","rating":3,"version":` + strconv.Itoa(version) + `}`
}

func TestUpdateBook(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: 更新した本を返し、版が進む", func(t *testing.T) {
		isbn := nextISBN(t)
		b := mustCreateBook(t, db, isbn)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/" + itoa(int64(b.ID)), Cookie: apitest.Session, Body: updateBody(b.Version)})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.BookResponse](t, rec)
		if got.Summary != "読み返した要約" || got.Comment != "読み返した感想" || got.Rating != 3 || got.Version != b.Version+1 {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: 存在しない分野タグは tagIdsNotFound で、本は変わらない", func(t *testing.T) {
		isbn := nextISBN(t)
		b := mustCreateBook(t, db, isbn)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/" + itoa(int64(b.ID)), Cookie: apitest.Session,
			Body: `{"summary":"要約","tagIds":[999999999],"comment":"良書","rating":3,"version":` + strconv.Itoa(b.Version) + `}`})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("tagIds", "tagIdsNotFound"))
		assertBookUnchanged(t, db, b)
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/1", Body: updateBody(1)})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/1", Cookie: apitest.Session, Origin: "http://evil.example", Body: updateBody(1)})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("404: 存在しない ID は bookNotFound", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/999999999", Cookie: apitest.Session, Body: updateBody(1)})
		apitest.AssertCode(t, rec, http.StatusNotFound, "bookNotFound")
	})

	t.Run("409: 版が一致しなければ versionConflict で、本は変わらない", func(t *testing.T) {
		isbn := nextISBN(t)
		b := mustCreateBook(t, db, isbn)
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, isbn)}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/" + itoa(int64(b.ID)), Cookie: apitest.Session, Body: updateBody(b.Version + 1)})
		apitest.AssertCode(t, rec, http.StatusConflict, "versionConflict")
		assertBookUnchanged(t, db, b)
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/1", Cookie: apitest.Session, Body: updateBody(1), Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

// TestUpdateBook_Constraints は、更新に固有の制約（path の id・version）を確かめる。本文の項目の制約は登録と同じスキーマの値を TestRegisterBook_Constraints で確かめている。
func TestUpdateBook_Constraints(t *testing.T) {
	db := connectIntegrationDB(t)
	s := newServer(t, db, &fakeCatalog{}, nil)
	tests := []struct {
		name, path, body, field, rule string
	}{
		{"id 0", "/api/admin/books/0", updateBody(1), "id", "minimum"},
		{"id 数字以外", "/api/admin/books/abc", updateBody(1), "id", "type"},
		{"version 欠落", "/api/admin/books/1", `{"summary":"要約","comment":"良書","rating":3}`, "version", "required"},
		{"summary 101 文字", "/api/admin/books/1", `{"summary":"` + strings.Repeat("あ", 101) + `","comment":"良書","rating":3,"version":1}`, "summary", "maxLength"},
		{"comment 5001 文字", "/api/admin/books/1", `{"summary":"要約","comment":"` + strings.Repeat("あ", 5001) + `","rating":3,"version":1}`, "comment", "maxLength"},
		{"rating 6", "/api/admin/books/1", `{"summary":"要約","comment":"良書","rating":6,"version":1}`, "rating", "maximum"},
	}
	for _, tt := range tests {
		t.Run("弾く: "+tt.name, func(t *testing.T) {
			rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: tt.path, Cookie: apitest.Session, Body: tt.body})
			apitest.AssertFieldErrors(t, rec, apitest.FieldError(tt.field, tt.rule))
		})
	}
	t.Run("通す: id 1（存在しなければ 404）", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodPut, Path: "/api/admin/books/999999999", Cookie: apitest.Session, Body: updateBody(1)})
		apitest.AssertNoFieldError(t, rec, "id", "minimum")
	})
}

// assertBookUnchanged は、失敗したはずの更新の後で本の内容と版が呼び出し前のままであることを確かめる。
func assertBookUnchanged(t *testing.T, db *gorm.DB, before *bookdomain.Book) {
	t.Helper()
	after, err := pgbook.NewRepository(db).FindByID(context.Background(), before.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if after.Comment.String() != before.Comment.String() || after.Summary.String() != before.Summary.String() ||
		after.Rating.Int() != before.Rating.Int() || after.Version != before.Version {
		t.Fatalf("book changed on a failed update: %+v", after)
	}
}

func TestDeleteBook(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("204: 削除すると DB から消える", func(t *testing.T) {
		b := mustCreateBook(t, db, nextISBN(t))
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/" + itoa(int64(b.ID)), Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusNoContent)
		var count int64
		db.Table("book").Where("id = ?", int64(b.ID)).Count(&count)
		if count != 0 {
			t.Fatal("the book must be deleted")
		}
	})

	t.Run("400: id が 0", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/0", Cookie: apitest.Session})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("id", "minimum"))
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/1"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/1", Cookie: apitest.Session, Origin: "http://evil.example"})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("404: 存在しない ID は bookNotFound", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/999999999", Cookie: apitest.Session})
		apitest.AssertCode(t, rec, http.StatusNotFound, "bookNotFound")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodDelete, Path: "/api/admin/books/1", Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestLookupCatalog(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: 外部カタログの書誌を返す", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{entry: catalogEntry(t, "9784873118703")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/978-4-87311-870-3", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.CatalogResponse](t, rec)
		if got.Isbn != "9784873118703" || got.Title != "結合テストの書名" || got.Authors != "結合テスト著者" || got.AmazonURL == nil || got.CoverURL != nil {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: ISBN が 18 文字なら maxLength", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/978-4-87311-870-33", Cookie: apitest.Session})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("isbn", "maxLength"))
	})

	t.Run("通す: ISBN 17 文字（形式が合わなければ errors の無い 400）", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/978-4-87311-870-0", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusBadRequest)
		apitest.AssertNoFieldError(t, rec, "isbn", "maxLength")
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/9784873118703"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("404: 外部カタログに無ければ catalogEntryNotFound", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{err: domainerr.New(domainerr.ErrNotFound, "catalogEntryNotFound", "無い")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/9784873118703", Cookie: apitest.Session})
		apitest.AssertCode(t, rec, http.StatusNotFound, "catalogEntryNotFound")
	})

	t.Run("500: 外部カタログの障害", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{err: errors.New("openbd down")}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/catalog/9784873118703", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestSearchBooks(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: 登録した本が一覧に載る", func(t *testing.T) {
		b := mustCreateBook(t, db, nextISBN(t))
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books?limit=100", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.SearchBooksResponse](t, rec)
		found := false
		for _, it := range got.Items {
			if it.ID == int64(b.ID) {
				found = it.Isbn == b.ISBN.String() && it.Title == "結合テストの書名" && it.Summary == "結合テスト用の一言まとめ" && it.Rating == 4
			}
		}
		if !found || got.Limit != 100 || got.Offset != 0 {
			t.Fatalf("book %d not listed as expected: %+v", b.ID, got)
		}
	})

	t.Run("400: limit が 0", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books?limit=0", Cookie: apitest.Session})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("limit", "minimum"))
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books", Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestReadBook(t *testing.T) {
	db := connectIntegrationDB(t)

	t.Run("200: 本の詳細と版を返す", func(t *testing.T) {
		b := mustCreateBook(t, db, nextISBN(t))
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books/" + itoa(int64(b.ID)), Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.BookResponse](t, rec)
		if got.ID != int64(b.ID) || got.Isbn != b.ISBN.String() || got.Comment != "結合テスト用の感想" || got.Version != 1 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("400: id が 0", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books/0", Cookie: apitest.Session})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("id", "minimum"))
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books/1"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("404: 存在しない ID は bookNotFound", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books/999999999", Cookie: apitest.Session})
		apitest.AssertCode(t, rec, http.StatusNotFound, "bookNotFound")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		s := newServer(t, db, &fakeCatalog{}, nil)
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/books/1", Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}
