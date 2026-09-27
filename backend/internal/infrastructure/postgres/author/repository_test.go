package author_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	pgauthor "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/author"
)

func newAuthor(t *testing.T, name string, birthDate *time.Time) *domainauthor.Author {
	t.Helper()
	n, err := domainauthor.NewName(name)
	if err != nil {
		t.Fatal(err)
	}
	var b *domainauthor.BirthDate
	if birthDate != nil {
		v := domainauthor.RestoreBirthDate(*birthDate)
		b = &v
	}
	return domainauthor.New(n, b)
}

func createAuthor(t *testing.T, db *gorm.DB, repo domainauthor.Repository, a *domainauthor.Author) {
	t.Helper()
	if err := repo.Create(context.Background(), a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM author WHERE id = ?", int64(a.ID)) })
}

func TestRepository_CreateThenFindByID(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauthor.NewRepository(db)
	birth := time.Date(1909, 6, 19, 0, 0, 0, 0, time.UTC)
	a := newAuthor(t, "repo-test-太宰治", &birth)

	createAuthor(t, db, repo, a)
	if a.ID == 0 || a.Version != 1 {
		t.Fatalf("Create set ID=%d Version=%d, want a new ID and version 1", a.ID, a.Version)
	}

	got, err := repo.FindByID(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name.String() != "repo-test-太宰治" || got.BirthDate == nil || !got.BirthDate.Time().Equal(birth) || got.Version != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := pgauthor.NewRepository(connectTestDB(t))

	if _, err := repo.FindByID(context.Background(), domainauthor.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestRepository_CountByIDs_CountsOnlyExistingAuthors(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauthor.NewRepository(db)
	a := newAuthor(t, "repo-test-count", nil)
	createAuthor(t, db, repo, a)

	n, err := repo.CountByIDs(context.Background(), []domainauthor.ID{a.ID, -1})
	if err != nil {
		t.Fatalf("CountByIDs: %v", err)
	}
	if n != 1 {
		t.Fatalf("count = %d, want 1", n)
	}
	if n, err := repo.CountByIDs(context.Background(), nil); err != nil || n != 0 {
		t.Fatalf("CountByIDs(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

func TestRepository_Update(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauthor.NewRepository(db)
	birth := time.Date(1909, 6, 19, 0, 0, 0, 0, time.UTC)
	a := newAuthor(t, "repo-test-before", &birth)
	createAuthor(t, db, repo, a)

	t.Run("バージョン一致なら更新しバージョンが進む", func(t *testing.T) {
		updated := newAuthor(t, "repo-test-after", nil)
		updated.ID, updated.Version = a.ID, 1
		if err := repo.Update(context.Background(), updated); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Version != 2 {
			t.Fatalf("Version = %d, want 2", updated.Version)
		}
		got, err := repo.FindByID(context.Background(), a.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Name.String() != "repo-test-after" || got.BirthDate != nil || got.Version != 2 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("古いバージョンはConflictで行は変わらない", func(t *testing.T) {
		stale := newAuthor(t, "repo-test-stale", nil)
		stale.ID, stale.Version = a.ID, 1
		if err := repo.Update(context.Background(), stale); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if stale.Version != 1 {
			t.Fatalf("Version = %d, want unchanged 1", stale.Version)
		}
		got, err := repo.FindByID(context.Background(), a.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Name.String() != "repo-test-after" || got.Version != 2 {
			t.Fatalf("row changed on conflict: %+v", got)
		}
	})
}
