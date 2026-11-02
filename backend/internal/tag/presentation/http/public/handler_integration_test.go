package public_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"

	publicapi "github.com/mrstsgk/book-management-system/backend/api/gen/public"
	"github.com/mrstsgk/book-management-system/backend/internal/apitest"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagpublic "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/public"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// connectTestDB はローカルの開発用 DB に接続する。繋がらなければ skip せず失敗にする（docs/rules/testing.md）。
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgplatform.Connect(pgplatform.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up migrate-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM tags LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

type testServer struct {
	apitest.PublicUnimplemented
	*tagpublic.Handler
}

func newServer(t *testing.T, db *gorm.DB) *apitest.Server {
	t.Helper()
	q := pgtag.NewQuery(db)
	return apitest.NewPublicServer(t, &testServer{Handler: &tagpublic.Handler{
		SearchTagsUsecase: &tagqry.SearchTagsUsecaseImpl{Tags: q},
		CountBooksUsecase: &tagqry.CountBooksUsecaseImpl{Tags: q},
	}})
}

// uniqueTagName は実行のたびに異なる、上限（30 文字）に収まるタグ名を返す。
func uniqueTagName(t *testing.T, label string) string {
	t.Helper()
	return fmt.Sprintf("pt-%s-%09d", label, time.Now().UnixNano()%1_000_000_000)
}

// mustCreateTag はタグを DB に直接 1 件入れる（追加の API は管理側の Handler の担当のため）。
func mustCreateTag(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	var id int64
	if err := db.Raw("INSERT INTO tags (name, version) VALUES (?, 1) RETURNING id", name).Scan(&id).Error; err != nil {
		t.Fatalf("setup: insert tag %q: %v", name, err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM tags WHERE id = ?", id) })
	return id
}

// mustCreateTaggedBook は本を DB に直接 1 件入れ、tagID を付ける。
func mustCreateTaggedBook(t *testing.T, db *gorm.DB, tagID int64) {
	t.Helper()
	twelve := fmt.Sprintf("978%09d", time.Now().UnixNano()%1_000_000_000)
	sum := 0
	for i, d := range twelve {
		w := 1
		if i%2 == 1 {
			w = 3
		}
		sum += int(d-'0') * w
	}
	isbn := twelve + string(byte((10-sum%10)%10)+'0')
	var bookID int64
	if err := db.Raw(
		"INSERT INTO books (isbn, title, summary, comment, rating, version) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		isbn, "pres-tag-test-書名", "pres-tag-test-まとめ", "pres-tag-test-感想", 5, 1,
	).Scan(&bookID).Error; err != nil {
		t.Fatalf("setup: insert book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM books WHERE id = ?", bookID) })
	if err := db.Exec("INSERT INTO book_tags (book_id, tag_id) VALUES (?, ?)", bookID, tagID).Error; err != nil {
		t.Fatalf("setup: insert book_tags: %v", err)
	}
}

func TestSearchTags(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 登録したタグを全項目で返す", func(t *testing.T) {
		name := uniqueTagName(t, "list")
		id := mustCreateTag(t, db, name)
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/tags"})
		apitest.AssertStatus(t, rec, http.StatusOK)
		for _, it := range apitest.Decode[publicapi.SearchTagsResponse](t, rec).Items {
			if it.ID == id {
				if it.Name != name || it.Version != 1 {
					t.Fatalf("item = %+v", it)
				}
				return
			}
		}
		t.Fatalf("tag %d not in the response", id)
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/tags", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestCountBooks(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 本が付いたタグだけを冊数付きで返す", func(t *testing.T) {
		taggedName := uniqueTagName(t, "cnt")
		tagged := mustCreateTag(t, db, taggedName)
		untagged := mustCreateTag(t, db, uniqueTagName(t, "zero"))
		mustCreateTaggedBook(t, db, tagged)

		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/tags/counts"})
		apitest.AssertStatus(t, rec, http.StatusOK)
		byID := map[int64]publicapi.TagBookCountResponse{}
		for _, it := range apitest.Decode[publicapi.CountBooksResponse](t, rec).Items {
			byID[it.ID] = it
		}
		if it, ok := byID[tagged]; !ok || it.BookCount != 1 || it.Name != taggedName {
			t.Fatalf("tagged item = %+v, want bookCount=1", it)
		}
		if _, ok := byID[untagged]; ok {
			t.Fatal("a tag with no books must be excluded")
		}
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/tags/counts", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}
