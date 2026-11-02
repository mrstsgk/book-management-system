package admin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagadmin "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/admin"
	tagpublic "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/public"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
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
func newIntegrationHandler(db *gorm.DB) *tagadmin.Handler {
	repo := pgtag.NewRepository(db)
	q := pgtag.NewQuery(db)
	return &tagadmin.Handler{
		RegisterTagUsecase: &tagcmd.RegisterTagUsecaseImpl{Tags: repo, TagQuery: q},
		RenameTagUsecase:   &tagcmd.RenameTagUsecaseImpl{Tags: repo, TagQuery: q},
		DeleteTagUsecase:   &tagcmd.DeleteTagUsecaseImpl{Tags: repo},
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

// uniqueTagName は実行のたびに異なるタグ名を返す（固定名だと、共有DBに前回のテスト行が
// 残っている場合や同時実行で一意制約に落ちるため）。タグ名の上限（30文字）に収まるよう、
// 接頭辞と連番を短くしている。
func uniqueTagName(t *testing.T, label string) string {
	t.Helper()
	return fmt.Sprintf("pt-%s-%06d", label, time.Now().UnixNano()%1_000_000)
}

func TestIntegration_RegisterTag_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	name := uniqueTagName(t, "登録")
	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, name), true)
	got := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", got.ID) })
	if got.Name != name || got.ID == 0 || got.Version != 1 {
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

	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "無認証")), false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_RegisterTag_400_Duplicated(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	body := fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "重複"))
	first := serve(t, h, http.MethodPost, "/api/tags", body, true)
	firstTag := decodeOK[tagpublic.Response](t, first, http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", firstTag.ID) })

	rec := serve(t, h, http.MethodPost, "/api/tags", body, true)
	assertFieldErrors(t, rec, []httpserver.FieldError{{Field: "name", Rule: "tagNameDuplicated"}})
}

func TestIntegration_Rename_400_Duplicated(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	otherName := uniqueTagName(t, "改名先の既存")
	other := decodeOK[tagpublic.Response](t, serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, otherName), true), http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", other.ID) })
	created := decodeOK[tagpublic.Response](t, serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "改名元")), true), http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec := serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID),
		fmt.Sprintf(`{"name":%q,"version":%d}`, otherName, created.Version), true)
	assertFieldErrors(t, rec, []httpserver.FieldError{{Field: "name", Rule: "tagNameDuplicated"}})
	var got struct{ Name string }
	db.Table("tag").Select("name").Where("id = ?", created.ID).Scan(&got)
	if got.Name == otherName {
		t.Fatal("the tag must not be renamed when the name is duplicated")
	}
}

// assertFieldErrors は 400 と errors[{field, rule}] の内容を確かめる。
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

func TestIntegration_Rename_200(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	afterName := uniqueTagName(t, "改名後")
	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "改名前")), true)
	created := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID),
		fmt.Sprintf(`{"name":%q,"version":%d}`, afterName, created.Version), true)
	got := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	if got.Name != afterName || got.Version != created.Version+1 {
		t.Fatalf("got %+v", got)
	}
}

func TestIntegration_Rename_400(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "不正入力対象")), true)
	created := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID), `{"name":""}`, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_401(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPut, "/api/tags/1", fmt.Sprintf(`{"name":%q,"version":1}`, uniqueTagName(t, "無認証改名")), false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_404(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPut, "/api/tags/999999999", fmt.Sprintf(`{"name":%q,"version":1}`, uniqueTagName(t, "存在しない")), true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Rename_409(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "競合対象")), true)
	created := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodPut, fmt.Sprintf("/api/tags/%d", created.ID),
		fmt.Sprintf(`{"name":%q,"version":%d}`, uniqueTagName(t, "競合後"), created.Version+1), true)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestIntegration_Delete_204(t *testing.T) {
	db := connectTestDB(t)
	h := newIntegrationHandler(db)

	rec := serve(t, h, http.MethodPost, "/api/tags", fmt.Sprintf(`{"name":%q}`, uniqueTagName(t, "削除対象")), true)
	created := decodeOK[tagpublic.Response](t, rec, http.StatusOK)
	// 本題のDELETEアサーションが失敗して先に t.Fatalf で止まっても、この行は残さない。
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", created.ID) })

	rec = serve(t, h, http.MethodDelete, fmt.Sprintf("/api/tags/%d", created.ID), "", true)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
	}

	// 一覧の API は閲覧側の Handler の担当なので、消えたことは DB で確かめる
	var n int64
	if err := db.Raw("SELECT COUNT(*) FROM tag WHERE id = ?", created.ID).Scan(&n).Error; err != nil {
		t.Fatalf("count tag: %v", err)
	}
	if n != 0 {
		t.Fatalf("tag %d still present after delete", created.ID)
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
