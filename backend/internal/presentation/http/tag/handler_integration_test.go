package tag_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"

	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

// Presentation → Infrastructure の結合テスト（docs/rules/testing.md）。
// 実際の Repository/Query（実 PostgreSQL）を注入した Handler に、実 HTTP リクエストを通す。
// 分岐網羅はしない（既存の Fake ベースの handler_test.go・usecase の Fake テスト・
// postgres/tag の契約テストが担う）。ここでは各エンドポイントが返しうる HTTP ステータスごとに
// 最低1件、配線が噛み合っていることだけを確かめる。

// connectTestDB はローカルの開発用 DB に接続する。契約テストと同じ理由で、繋がらなければ skip する。
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgcommon.Connect(pgcommon.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up migrate-up` first): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM tag LIMIT 0").Error; err != nil {
		t.Skipf("skipping: schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// newIntegrationHandler は cmd/api/main.go と同じ形（実 Repository/Query を注入）で Handler を組み立てる。
func newIntegrationHandler(db *gorm.DB) *httptag.Handler {
	repo := pgtag.NewRepository(db)
	q := pgtag.NewQuery(db)
	return &httptag.Handler{
		RegisterUC:   &tagcmd.RegisterUsecaseImpl{Tags: repo},
		RenameUC:     &tagcmd.RenameUsecaseImpl{Tags: repo},
		DeleteUC:     &tagcmd.DeleteUsecaseImpl{Tags: repo},
		ListUC:       &tagqry.ListUsecaseImpl{Tags: q},
		CountBooksUC: &tagqry.CountBooksUsecaseImpl{Tags: q},
	}
}

func decodeIntegration[T any](t *testing.T, body []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decode %s: %v", body, err)
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

func TestIntegration_List_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-一覧"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup: register status = %d, body=%s", rec.Code, rec.Body.String())
	}
	created := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodGet, "/api/tags", "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	got := decodeIntegration[httptag.ListResponse](t, rec.Body.Bytes())
	found := false
	for _, it := range got.Items {
		if it.Name == "pres-tag-test-一覧" {
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

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-集計あり"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup: register (tagged) status = %d, body=%s", rec.Code, rec.Body.String())
	}
	tagged := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", tagged.ID) })

	rec = serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-0冊"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup: register (untagged) status = %d, body=%s", rec.Code, rec.Body.String())
	}
	untagged := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
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

	rec = serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	got := decodeIntegration[httptag.BookCountListResponse](t, rec.Body.Bytes())
	byID := map[int64]httptag.BookCountResponse{}
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

func TestIntegration_RegisterTag_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-登録"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	got := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", got.ID) })
	if got.Name != "pres-tag-test-登録" || got.ID == 0 || got.Version != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestIntegration_RegisterTag_400(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{}`, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_RegisterTag_401(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-無認証"}`, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_RegisterTag_409(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	body := `{"name":"pres-tag-test-重複"}`
	first := serve(t, h, http.MethodPost, "/api/tags", body, true)
	if first.Code != http.StatusOK {
		t.Fatalf("setup: first register status = %d, body=%s", first.Code, first.Body.String())
	}
	firstTag := decodeIntegration[httptag.Response](t, first.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", firstTag.ID) })

	rec := serve(t, h, http.MethodPost, "/api/tags", body, true)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-改名前"}`, true)
	created := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID),
		fmt.Sprintf(`{"name":"pres-tag-test-改名後","version":%d}`, created.Version), true)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	got := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	if got.Name != "pres-tag-test-改名後" || got.Version != created.Version+1 {
		t.Fatalf("got %+v", got)
	}
}

func TestIntegration_Rename_400(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-不正入力対象"}`, true)
	created := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID), `{"name":""}`, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_401(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPut, "/api/tags/1", `{"name":"pres-tag-test-無認証改名","version":1}`, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_404(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPut, "/api/tags/999999999", `{"name":"pres-tag-test-存在しない","version":1}`, true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_409(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-競合対象"}`, true)
	created := decodeIntegration[httptag.Response](t, rec.Body.Bytes())
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID),
		fmt.Sprintf(`{"name":"pres-tag-test-競合後","version":%d}`, created.Version+1), true)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Delete_204(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"pres-tag-test-削除対象"}`, true)
	created := decodeIntegration[httptag.Response](t, rec.Body.Bytes())

	rec = serve(t, h, http.MethodDelete, fmt.Sprintf("/api/tags/%d", created.ID), "", true)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
	}

	rec = serve(t, h, http.MethodGet, "/api/tags", "", false)
	got := decodeIntegration[httptag.ListResponse](t, rec.Body.Bytes())
	for _, it := range got.Items {
		if it.ID == created.ID {
			t.Fatalf("tag %d still present after delete", created.ID)
		}
	}
}

func TestIntegration_Delete_401(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodDelete, "/api/tags/1", "", false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Delete_404(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodDelete, "/api/tags/999999999", "", true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
	}
}
