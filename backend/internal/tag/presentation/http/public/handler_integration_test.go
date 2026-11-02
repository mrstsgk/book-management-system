package public_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagpublic "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/public"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// Presentation → Infrastructure の結合テスト（docs/rules/testing.md）。
// 実際の Repository/Query（実 PostgreSQL）を注入した Handler に、実 HTTP リクエストを通す。
// 分岐網羅はしない（既存の Fake ベースの handler_test.go・usecase の Fake テスト・
// postgres/tag の契約テストが担う）。ここでは各エンドポイントが返しうる HTTP ステータスごとに
// 最低1件、配線が噛み合っていることだけを確かめる。

// connectTestDB はローカルの開発用 DB に接続する。契約テストと同じ理由で、繋がらなければ skip せず失敗にする。
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgplatform.Connect(pgplatform.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up migrate-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM tag LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// newIntegrationHandler は cmd/api/main.go と同じ形（実 Repository/Query を注入）で Handler を組み立てる。
func newIntegrationHandler(db *gorm.DB) *tagpublic.Handler {
	q := pgtag.NewQuery(db)
	return &tagpublic.Handler{
		SearchTagsUsecase: &tagqry.SearchTagsUsecaseImpl{Tags: q},
		CountBooksUsecase: &tagqry.CountBooksUsecaseImpl{Tags: q},
	}
}

// decodeOK はステータスを先に確かめてからデコードする（エラー応答はレスポンスの形が違うため、
// ステータスを見ずにデコードすると想定外のエラーを見逃して成功したかのように扱ってしまう）。
func decodeOK[T any](t *testing.T, rec *httptest.ResponseRecorder, want int) T {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, want, rec.Body.String())
	}
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.Bytes(), err)
	}
	return v
}

// isbn13CheckDigit は internal/domain/book.isbn13CheckDigit と同じ計算（重み1/3の交互和）。
// 非公開なのでこのテストからは呼べず、同じアルゴリズムをここに複製している。
func isbn13CheckDigit(twelve string) byte {
	sum := 0
	for i, d := range twelve {
		w := 1
		if i%2 == 1 {
			w = 3
		}
		sum += int(d-'0') * w
	}
	return byte((10-sum%10)%10) + '0'
}

// mustUniqueISBN は実行のたびに異なる、チェックディジットの正しい ISBN-13 を返す
// （固定のISBN文字列だと、共有DBに前回のテスト行が残っている場合や同時実行で一意制約に落ちるため）。
func mustUniqueISBN(t *testing.T) string {
	t.Helper()
	twelve := fmt.Sprintf("978%09d", time.Now().UnixNano()%1_000_000_000)
	return twelve + string(isbn13CheckDigit(twelve))
}

// uniqueTagName は実行のたびに異なるタグ名を返す（固定名だと、共有DBに前回のテスト行が
// 残っている場合や同時実行で一意制約に落ちるため）。タグ名の上限（30文字）に収まるよう、
// 接頭辞と連番を短くしている。
func uniqueTagName(t *testing.T, label string) string {
	t.Helper()
	return fmt.Sprintf("pt-%s-%06d", label, time.Now().UnixNano()%1_000_000)
}

func TestIntegration_List_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	name := uniqueTagName(t, "一覧")
	created := mustCreateTag(t, db, name)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec := serve(t, h, http.MethodGet, "/api/tags", "", false)
	got := decodeOK[tagpublic.ListResponse](t, rec, http.StatusOK)
	found := false
	for _, it := range got.Items {
		if it.Name == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("items = %+v, want to contain the registered tag", got.Items)
	}
}

func TestIntegration_CountBooks_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	tagged := mustCreateTag(t, db, uniqueTagName(t, "集計あり"))
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", tagged.ID) })

	untagged := mustCreateTag(t, db, uniqueTagName(t, "0冊"))
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", untagged.ID) })

	bookISBN := mustUniqueISBN(t)
	var bookID int64
	if err := db.Raw(
		"INSERT INTO book (isbn, title, summary, comment, rating, version) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		bookISBN, "pres-tag-test-書名", "pres-tag-test-まとめ", "pres-tag-test-感想", 5, 1,
	).Scan(&bookID).Error; err != nil {
		t.Fatalf("setup: insert book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", bookID) })
	if err := db.Exec("INSERT INTO book_tag (book_id, tag_id) VALUES (?, ?)", bookID, tagged.ID).Error; err != nil {
		t.Fatalf("setup: insert book_tag: %v", err)
	}

	rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
	got := decodeOK[tagpublic.BookCountListResponse](t, rec, http.StatusOK)
	byID := map[int64]tagpublic.BookCountResponse{}
	for _, it := range got.Items {
		byID[it.ID] = it
	}
	if it, ok := byID[tagged.ID]; !ok || it.BookCount != 1 {
		t.Fatalf("tagged item = %+v, want bookCount=1", it)
	}
	if _, ok := byID[untagged.ID]; ok {
		t.Fatalf("items = %+v, want the 0-book tag excluded", got.Items)
	}
}

// mustCreateTag はテスト用のタグを DB に直接 1 件入れる（追加の API は管理側の Handler の担当で、
// 閲覧側のテストからは使えないため）。
func mustCreateTag(t *testing.T, db *gorm.DB, name string) tagpublic.Response {
	t.Helper()
	var id int64
	if err := db.Raw("INSERT INTO tag (name, version) VALUES (?, 1) RETURNING id", name).Scan(&id).Error; err != nil {
		t.Fatalf("setup: insert tag %q: %v", name, err)
	}
	return tagpublic.Response{ID: id, Name: name, Version: 1}
}
