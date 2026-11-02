package admin_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	"github.com/mrstsgk/book-management-system/backend/internal/apitest"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagadmin "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/admin"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
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
	if err := db.Exec("SELECT 1 FROM tag LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// skipDuplicateCheck は、タグ名の重複の事前検査をすり抜ける Query。同時の追加・改名で DB の一意制約（最後の砦）に当たる状況を作る。
type skipDuplicateCheck struct{ tagdomain.Query }

func (skipDuplicateCheck) ExistsByName(context.Context, string, tagdomain.ID) (bool, error) {
	return false, nil
}

type testServer struct {
	apitest.AdminUnimplemented
	*tagadmin.Handler
}

// newServer は、本物の Repository/Query/UseCase を実 DB に配線した管理の Handler をサーバーに載せる。
func newServer(t *testing.T, db *gorm.DB, q tagdomain.Query) *apitest.Server {
	t.Helper()
	repo := pgtag.NewRepository(db)
	if q == nil {
		q = pgtag.NewQuery(db)
	}
	return apitest.NewAdminServer(t, &testServer{Handler: &tagadmin.Handler{
		SearchTagsUsecase:  &tagqry.SearchTagsUsecaseImpl{Tags: q},
		CountBooksUsecase:  &tagqry.CountBooksUsecaseImpl{Tags: q},
		RegisterTagUsecase: &tagcmd.RegisterTagUsecaseImpl{Tags: repo, TagQuery: q},
		RenameTagUsecase:   &tagcmd.RenameTagUsecaseImpl{Tags: repo, TagQuery: q},
		DeleteTagUsecase:   &tagcmd.DeleteTagUsecaseImpl{Tags: repo},
	}}, apitest.AcceptSession{})
}

// uniqueTagName は実行のたびに異なる、上限（30 文字）に収まるタグ名を返す（共有 DB に残った行と一意制約で衝突しないため）。
func uniqueTagName(t *testing.T, label string) string {
	t.Helper()
	return fmt.Sprintf("it-%s-%09d", label, time.Now().UnixNano()%1_000_000_000)
}

// mustCreateTag は DB へ直接タグを作り、ID と版を返す。
func mustCreateTag(t *testing.T, db *gorm.DB, name string) (int64, int) {
	t.Helper()
	var row struct {
		ID      int64
		Version int
	}
	if err := db.Raw("INSERT INTO tag (name, version) VALUES (?, 1) RETURNING id, version", name).Scan(&row).Error; err != nil {
		t.Fatalf("insert tag: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", row.ID) })
	return row.ID, row.Version
}

func tagNameOf(t *testing.T, db *gorm.DB, id int64) string {
	t.Helper()
	var name string
	db.Raw("SELECT name FROM tag WHERE id = ?", id).Scan(&name)
	return name
}

func path(id int64) string { return "/api/admin/tags/" + strconv.FormatInt(id, 10) }

func TestRegisterTag(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 追加したタグを返す", func(t *testing.T) {
		name := uniqueTagName(t, "reg")
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: apitest.Session, Body: `{"name":"` + name + `"}`})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.TagResponse](t, rec)
		t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE id = ?", got.ID) })
		if got.Name != name || got.Version != 1 || got.ID < 1 {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: 同じ名前は tagNameDuplicated", func(t *testing.T) {
		name := uniqueTagName(t, "dup")
		mustCreateTag(t, db, name)
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: apitest.Session, Body: `{"name":"` + name + `"}`})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("name", "tagNameDuplicated"))
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Body: `{"name":"x"}`})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: apitest.Session, Origin: "http://evil.example", Body: `{"name":"x"}`})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("409: 事前の検査をすり抜けた同時追加は一意制約で tagNameDuplicated", func(t *testing.T) {
		name := uniqueTagName(t, "race")
		mustCreateTag(t, db, name)
		rec := newServer(t, db, skipDuplicateCheck{pgtag.NewQuery(db)}).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: apitest.Session, Body: `{"name":"` + name + `"}`})
		apitest.AssertCode(t, rec, http.StatusConflict, "tagNameDuplicated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPost, Path: "/api/admin/tags", Cookie: apitest.Session, Body: `{"name":"x"}`, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

// TestTagName_Constraints は、YAML に書いた name・version・id の制約 1 つにつき、通る値と弾かれる値を 1 組ずつ確かめる。
func TestTagName_Constraints(t *testing.T) {
	db := connectTestDB(t)
	s := newServer(t, db, nil)
	tests := []struct {
		name, method, path, body, field, rule string
		pass                                  bool
	}{
		{"追加: name 欠落は弾く", http.MethodPost, "/api/admin/tags", `{}`, "name", "required", false},
		{"追加: name 0 文字は弾く", http.MethodPost, "/api/admin/tags", `{"name":""}`, "name", "minLength", false},
		{"追加: name 31 文字は弾く", http.MethodPost, "/api/admin/tags", `{"name":"` + strings.Repeat("あ", 31) + `"}`, "name", "maxLength", false},
		{"追加: name 30 文字は通る", http.MethodPost, "/api/admin/tags", `{"name":"` + strings.Repeat("あ", 30) + `"}`, "name", "maxLength", true},
		{"改名: version 欠落は弾く", http.MethodPut, "/api/admin/tags/1", `{"name":"x"}`, "version", "required", false},
		{"改名: name 31 文字は弾く", http.MethodPut, "/api/admin/tags/1", `{"name":"` + strings.Repeat("あ", 31) + `","version":1}`, "name", "maxLength", false},
		{"改名: id 0 は弾く", http.MethodPut, "/api/admin/tags/0", `{"name":"x","version":1}`, "id", "minimum", false},
		{"改名: id 1 は通る", http.MethodPut, "/api/admin/tags/999999999", `{"name":"x","version":1}`, "id", "minimum", true},
		{"削除: id 0 は弾く", http.MethodDelete, "/api/admin/tags/0", "", "id", "minimum", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := s.Do(t, apitest.Request{Method: tt.method, Path: tt.path, Cookie: apitest.Session, Body: tt.body})
			if tt.pass {
				apitest.AssertNoFieldError(t, rec, tt.field, tt.rule)
				if rec.Code == http.StatusOK {
					db.Exec("DELETE FROM tag WHERE name = ?", strings.Repeat("あ", 30))
				}
				return
			}
			apitest.AssertFieldErrors(t, rec, apitest.FieldError(tt.field, tt.rule))
		})
	}
}

func TestRenameTag(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 改名したタグを返し、版が進む", func(t *testing.T) {
		id, version := mustCreateTag(t, db, uniqueTagName(t, "old"))
		newName := uniqueTagName(t, "new")
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(id), Cookie: apitest.Session, Body: `{"name":"` + newName + `","version":` + strconv.Itoa(version) + `}`})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[adminapi.TagResponse](t, rec)
		if got.ID != id || got.Name != newName || got.Version != version+1 {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: 自分以外と同じ名前は tagNameDuplicated で、名前は変わらない", func(t *testing.T) {
		taken := uniqueTagName(t, "taken")
		mustCreateTag(t, db, taken)
		before := uniqueTagName(t, "keep")
		id, version := mustCreateTag(t, db, before)
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(id), Cookie: apitest.Session, Body: `{"name":"` + taken + `","version":` + strconv.Itoa(version) + `}`})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("name", "tagNameDuplicated"))
		if got := tagNameOf(t, db, id); got != before {
			t.Fatalf("name changed to %q on a failed rename", got)
		}
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(1), Body: `{"name":"x","version":1}`})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(1), Cookie: apitest.Session, Origin: "http://evil.example", Body: `{"name":"x","version":1}`})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("404: 存在しない ID は tagNotFound", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(999999999), Cookie: apitest.Session, Body: `{"name":"x","version":1}`})
		apitest.AssertCode(t, rec, http.StatusNotFound, "tagNotFound")
	})

	t.Run("409: 版が一致しなければ versionConflict で、名前は変わらない", func(t *testing.T) {
		before := uniqueTagName(t, "ver")
		id, version := mustCreateTag(t, db, before)
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(id), Cookie: apitest.Session, Body: `{"name":"` + uniqueTagName(t, "x") + `","version":` + strconv.Itoa(version+1) + `}`})
		apitest.AssertCode(t, rec, http.StatusConflict, "versionConflict")
		if got := tagNameOf(t, db, id); got != before {
			t.Fatalf("name changed to %q on a failed rename", got)
		}
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodPut, Path: path(1), Cookie: apitest.Session, Body: `{"name":"x","version":1}`, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestDeleteTag(t *testing.T) {
	db := connectTestDB(t)

	t.Run("204: 削除すると DB から消える", func(t *testing.T) {
		id, _ := mustCreateTag(t, db, uniqueTagName(t, "del"))
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodDelete, Path: path(id), Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusNoContent)
		var count int64
		db.Table("tag").Where("id = ?", id).Count(&count)
		if count != 0 {
			t.Fatal("the tag must be deleted")
		}
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodDelete, Path: path(1)})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("403: Origin が違う", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodDelete, Path: path(1), Cookie: apitest.Session, Origin: "http://evil.example"})
		apitest.AssertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("404: 存在しない ID は tagNotFound", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodDelete, Path: path(999999999), Cookie: apitest.Session})
		apitest.AssertCode(t, rec, http.StatusNotFound, "tagNotFound")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodDelete, Path: path(1), Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestSearchTags(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 追加したタグが一覧に載る", func(t *testing.T) {
		name := uniqueTagName(t, "search")
		id, _ := mustCreateTag(t, db, name)
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusOK)
		found := false
		for _, it := range apitest.Decode[adminapi.SearchTagsResponse](t, rec).Items {
			if it.ID == id {
				found = it.Name == name && it.Version == 1
			}
		}
		if !found {
			t.Fatalf("tag %d (%s) not listed", id, name)
		}
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags", Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestCountBooks(t *testing.T) {
	db := connectTestDB(t)

	t.Run("200: 本の付いたタグだけを冊数付きで返す", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags/counts", Cookie: apitest.Session})
		apitest.AssertStatus(t, rec, http.StatusOK)
		for _, it := range apitest.Decode[adminapi.CountBooksResponse](t, rec).Items {
			if it.BookCount < 1 {
				t.Fatalf("tags without books must not be counted: %+v", it)
			}
		}
	})

	t.Run("401: 未ログイン", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags/counts"})
		apitest.AssertCode(t, rec, http.StatusUnauthorized, "unauthenticated")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := newServer(t, db, nil).Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/admin/tags/counts", Cookie: apitest.Session, Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}
