package book_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	pgauthor "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/author"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
)

func seedAuthor(t *testing.T, db *gorm.DB, name string, birthDate *time.Time) domainauthor.ID {
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
	a := domainauthor.New(n, b)
	if err := pgauthor.NewRepository(db).Create(context.Background(), a); err != nil {
		t.Fatalf("seed author: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM author WHERE id = ?", int64(a.ID)) })
	return a.ID
}

func newBook(t *testing.T, title string, price int64, authorIDs []domainauthor.ID, status domainbook.PublishStatus) *domainbook.Book {
	t.Helper()
	tt, err := domainbook.NewTitle(title)
	if err != nil {
		t.Fatal(err)
	}
	p, err := domainbook.NewPrice(price)
	if err != nil {
		t.Fatal(err)
	}
	b, err := domainbook.New(tt, p, authorIDs, status)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// cleanupBook is registered before authors are cleaned up (t.Cleanup is LIFO), so FKs never block it.
func cleanupBook(t *testing.T, db *gorm.DB, id func() domainbook.ID) {
	t.Helper()
	t.Cleanup(func() {
		db.Exec("DELETE FROM author_book WHERE book_id = ?", int64(id()))
		db.Exec("DELETE FROM book WHERE id = ?", int64(id()))
	})
}

func createBook(t *testing.T, db *gorm.DB, b *domainbook.Book) {
	t.Helper()
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("Create: %v", err)
	}
	cleanupBook(t, db, func() domainbook.ID { return b.ID })
}

func TestRepository_CreateThenFindByID(t *testing.T) {
	db := connectTestDB(t)
	a1 := seedAuthor(t, db, "repo-test-a1", nil)
	a2 := seedAuthor(t, db, "repo-test-a2", nil)
	repo := pgbook.NewRepository(db)

	tests := []struct {
		name  string
		price int64
	}{
		{name: "0円を往復できる", price: 0},
		{name: "上限の価格を往復できる", price: 99_999_999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBook(t, "repo-test-人間失格", tt.price, []domainauthor.ID{a2, a1}, domainbook.Published)
			createBook(t, db, b)
			if b.ID == 0 || b.Version != 1 {
				t.Fatalf("Create set ID=%d Version=%d, want a new ID and version 1", b.ID, b.Version)
			}

			got, err := repo.FindByID(context.Background(), b.ID)
			if err != nil {
				t.Fatalf("FindByID: %v", err)
			}
			if got.Title.String() != "repo-test-人間失格" || got.Price.Int64() != tt.price || got.Status != domainbook.Published || got.Version != 1 {
				t.Fatalf("got %+v", got)
			}
			if want := []domainauthor.ID{a1, a2}; !reflect.DeepEqual(got.AuthorIDs, want) {
				t.Fatalf("AuthorIDs = %v, want %v (sorted by id)", got.AuthorIDs, want)
			}
		})
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := pgbook.NewRepository(connectTestDB(t))

	if _, err := repo.FindByID(context.Background(), domainbook.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestRepository_FindByID_RejectsFractionalPrice(t *testing.T) {
	db := connectTestDB(t)
	a := seedAuthor(t, db, "repo-test-fraction", nil)
	var id int64
	if err := db.Raw("INSERT INTO book (title, price, publish_status, version) VALUES ('repo-test-fraction', 1.5, 1, 1) RETURNING id").Scan(&id).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	cleanupBook(t, db, func() domainbook.ID { return domainbook.ID(id) })
	db.Exec("INSERT INTO author_book (author_id, book_id, version) VALUES (?, ?, 1)", int64(a), id)

	if _, err := pgbook.NewRepository(db).FindByID(context.Background(), domainbook.ID(id)); err == nil {
		t.Fatal("expected an error instead of silently truncating a fractional price")
	}
}

func TestRepository_Create_UnknownAuthorLeavesNoBook(t *testing.T) {
	db := connectTestDB(t)
	b := newBook(t, "repo-test-orphan", 100, []domainauthor.ID{-1}, domainbook.Unpublished)

	if err := pgbook.NewRepository(db).Create(context.Background(), b); err == nil {
		t.Fatal("expected a foreign key error")
	}
	var n int64
	db.Raw("SELECT COUNT(*) FROM book WHERE title = 'repo-test-orphan'").Scan(&n)
	if n != 0 {
		db.Exec("DELETE FROM book WHERE title = 'repo-test-orphan'")
		t.Fatalf("book row was committed without its authors (%d rows)", n)
	}
	if b.ID != 0 || b.Version != 0 {
		t.Fatalf("b was mutated on failure: %+v", b)
	}
}

func TestRepository_Update(t *testing.T) {
	db := connectTestDB(t)
	a1 := seedAuthor(t, db, "repo-test-u1", nil)
	a2 := seedAuthor(t, db, "repo-test-u2", nil)
	repo := pgbook.NewRepository(db)
	b := newBook(t, "repo-test-before", 1500, []domainauthor.ID{a1}, domainbook.Unpublished)
	createBook(t, db, b)

	t.Run("バージョン一致なら内容と著者の関連を入れ替える", func(t *testing.T) {
		changed := newBook(t, "repo-test-after", 800, []domainauthor.ID{a2}, domainbook.Published)
		changed.ID, changed.Version = b.ID, 1
		if err := repo.Update(context.Background(), changed); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if changed.Version != 2 {
			t.Fatalf("Version = %d, want 2", changed.Version)
		}
		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Title.String() != "repo-test-after" || got.Price.Int64() != 800 || got.Status != domainbook.Published ||
			got.Version != 2 || !reflect.DeepEqual(got.AuthorIDs, []domainauthor.ID{a2}) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("古いバージョンはConflictで行も著者の関連も変わらない", func(t *testing.T) {
		stale := newBook(t, "repo-test-stale", 1, []domainauthor.ID{a1}, domainbook.Unpublished)
		stale.ID, stale.Version = b.ID, 1
		if err := repo.Update(context.Background(), stale); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if stale.Version != 1 {
			t.Fatalf("Version = %d, want unchanged 1", stale.Version)
		}
		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Title.String() != "repo-test-after" || got.Version != 2 || !reflect.DeepEqual(got.AuthorIDs, []domainauthor.ID{a2}) {
			t.Fatalf("row changed on conflict: %+v", got)
		}
	})
}

func TestRepository_AmazonURLAndImageKey(t *testing.T) {
	db := connectTestDB(t)
	a := seedAuthor(t, db, "repo-test-media", nil)
	repo := pgbook.NewRepository(db)
	ctx := context.Background()
	u, err := domainbook.NewAmazonURL("https://www.amazon.co.jp/dp/4101006059")
	if err != nil {
		t.Fatal(err)
	}
	key, err := domainbook.NewImageKey("books/test/cover.png")
	if err != nil {
		t.Fatal(err)
	}

	b := newBook(t, "repo-test-media", 100, []domainauthor.ID{a}, domainbook.Unpublished)
	b.ChangeAmazonURL(&u)
	b.ReplaceImage(key)
	createBook(t, db, b)

	got, err := repo.FindByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.AmazonURL == nil || *got.AmazonURL != u || got.ImageKey == nil || *got.ImageKey != key {
		t.Fatalf("got url=%v key=%v, want %v / %v", got.AmazonURL, got.ImageKey, u, key)
	}

	got.ChangeAmazonURL(nil)
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	cleared, err := repo.FindByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("FindByID after clearing: %v", err)
	}
	if cleared.AmazonURL != nil {
		t.Fatalf("AmazonURL = %v, want NULL after clearing", cleared.AmazonURL)
	}
	if cleared.ImageKey == nil || *cleared.ImageKey != key {
		t.Fatalf("ImageKey = %v, want it kept when only the URL is cleared", cleared.ImageKey)
	}

}
