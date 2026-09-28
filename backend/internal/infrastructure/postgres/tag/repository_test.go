package tag_test

import (
	"context"
	"errors"
	"testing"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
)

func mustName(t *testing.T, raw string) domaintag.Name {
	t.Helper()
	n, err := domaintag.NewName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func createTag(t *testing.T, repo domaintag.Repository, name string) *domaintag.Tag {
	t.Helper()
	tg := domaintag.New(mustName(t, name))
	if err := repo.Create(context.Background(), tg); err != nil {
		t.Fatalf("Create: %v", err)
	}
	return tg
}

func TestRepository_CreateThenFindByID(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'repo-test-%'") })

	tg := createTag(t, repo, "repo-test-データベース")
	if tg.ID == 0 || tg.Version != 1 {
		t.Fatalf("got ID=%d Version=%d, want a nonzero ID and Version 1", tg.ID, tg.Version)
	}

	got, err := repo.FindByID(context.Background(), tg.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name.String() != "repo-test-データベース" || got.Version != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := pgtag.NewRepository(connectTestDB(t))
	if _, err := repo.FindByID(context.Background(), domaintag.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestRepository_Create_DuplicateNameIsConflict(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'repo-test-%'") })
	createTag(t, repo, "repo-test-重複")

	dup := domaintag.New(mustName(t, "repo-test-重複"))
	if err := repo.Create(context.Background(), dup); !errors.Is(err, domaincommon.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestRepository_Update(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'repo-test-%'") })

	t.Run("バージョン一致なら名前を変えバージョンが進む", func(t *testing.T) {
		tg := createTag(t, repo, "repo-test-旧名")
		tg.Rename(mustName(t, "repo-test-新名"), tg.Version)
		if err := repo.Update(context.Background(), tg); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if tg.Version != 2 {
			t.Fatalf("Version = %d, want 2", tg.Version)
		}
		got, err := repo.FindByID(context.Background(), tg.ID)
		if err != nil || got.Name.String() != "repo-test-新名" {
			t.Fatalf("got %+v, err=%v", got, err)
		}
	})

	t.Run("古いバージョンはConflictで名前は変わらない", func(t *testing.T) {
		tg := createTag(t, repo, "repo-test-競合前")
		stale := domaintag.New(mustName(t, "repo-test-競合後"))
		stale.ID = tg.ID
		stale.Version = tg.Version - 1
		if err := repo.Update(context.Background(), stale); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		got, err := repo.FindByID(context.Background(), tg.ID)
		if err != nil || got.Name.String() != "repo-test-競合前" {
			t.Fatalf("row was mutated on failure: got %+v, err=%v", got, err)
		}
	})

	t.Run("改名先が既存の別タグと同名ならConflict", func(t *testing.T) {
		createTag(t, repo, "repo-test-既存名")
		tg := createTag(t, repo, "repo-test-改名したい")
		tg.Rename(mustName(t, "repo-test-既存名"), tg.Version)
		if err := repo.Update(context.Background(), tg); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}

func TestRepository_Delete(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	tg := createTag(t, repo, "repo-test-削除対象")

	if err := repo.Delete(context.Background(), tg.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), tg.ID); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("FindByID after delete: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(context.Background(), tg.ID); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("Delete again: err = %v, want ErrNotFound", err)
	}
}
