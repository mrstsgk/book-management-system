package book_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
)

// newBook はテスト用の読んだ本を作る。ISBN はテスト間で重ならないものを渡す。
func newBook(t *testing.T, isbn, title string, cover *domainbook.Cover, rating int, tagIDs ...domaintag.ID) *domainbook.Book {
	t.Helper()
	i, err := domainbook.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := domainbook.NewBibliography(title, "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	s, err := domainbook.NewSummary("一言まとめ")
	if err != nil {
		t.Fatal(err)
	}
	c, err := domainbook.NewComment("感想\n2行目")
	if err != nil {
		t.Fatal(err)
	}
	r, err := domainbook.NewRating(rating)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := domainbook.NewTagSelection(tagIDs)
	if err != nil {
		t.Fatal(err)
	}
	return domainbook.New(i, bib, cover, s, c, r, tags)
}

func mustCover(t *testing.T, url string) *domainbook.Cover {
	t.Helper()
	c, err := domainbook.NewCover(url, domainbook.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func createBook(t *testing.T, db *gorm.DB, b *domainbook.Book) {
	t.Helper()
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
}

func TestRepository_CreateThenFindByID(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)

	tests := []struct {
		name  string
		isbn  string
		cover *domainbook.Cover
	}{
		{name: "書影ありを往復できる", isbn: "9780000000002", cover: mustCover(t, "https://cover.openbd.jp/test.jpg")},
		{name: "書影なしを往復できる", isbn: "9780000000019", cover: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBook(t, tt.isbn, "repo-test-書名", tt.cover, 4)
			createBook(t, db, b)
			if b.ID == 0 || b.Version != 1 {
				t.Fatalf("Create set ID=%d Version=%d, want a new ID and version 1", b.ID, b.Version)
			}

			got, err := repo.FindByID(context.Background(), b.ID)
			if err != nil {
				t.Fatalf("FindByID: %v", err)
			}
			if got.ISBN != b.ISBN || got.Bibliography != b.Bibliography || got.Comment != b.Comment || got.Rating != b.Rating || got.Version != 1 {
				t.Fatalf("got %+v, want %+v", got, b)
			}
			if (got.Cover == nil) != (tt.cover == nil) || (got.Cover != nil && *got.Cover != *tt.cover) {
				t.Fatalf("Cover = %v, want %v", got.Cover, tt.cover)
			}
		})
	}
}

func TestRepository_Create_DuplicateISBNIsConflict(t *testing.T) {
	db := connectTestDB(t)
	first := newBook(t, "9780000000026", "repo-test-first", nil, 3)
	createBook(t, db, first)

	dup := newBook(t, "9780000000026", "repo-test-dup", nil, 3)
	if err := pgbook.NewRepository(db).Create(context.Background(), dup); !errors.Is(err, domaincommon.ErrConflict) {
		if dup.ID != 0 {
			db.Exec("DELETE FROM book WHERE id = ?", int64(dup.ID))
		}
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if dup.ID != 0 || dup.Version != 0 {
		t.Fatalf("dup was mutated on failure: ID=%d Version=%d", dup.ID, dup.Version)
	}
}

func TestRepository_Update(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)
	b := newBook(t, "9780000000033", "repo-test-before", mustCover(t, "https://cover.openbd.jp/before.jpg"), 2)
	createBook(t, db, b)

	t.Run("バージョン一致なら感想・評価・書誌・書影を更新しバージョンが進む", func(t *testing.T) {
		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatal(err)
		}
		comment, _ := domainbook.NewComment("読み返した")
		rating, _ := domainbook.NewRating(5)
		bib, _ := domainbook.NewBibliography("repo-test-after", "", "", "")
		got.ChangeReview(got.Summary, comment, rating, got.Tags, 1)
		got.RefreshCatalog(bib, nil)

		if err := repo.Update(context.Background(), got); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.Version != 2 {
			t.Fatalf("Version = %d, want 2", got.Version)
		}
		after, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatal(err)
		}
		if after.Comment != comment || after.Rating != rating || after.Bibliography != bib || after.Cover != nil || after.Version != 2 {
			t.Fatalf("got %+v", after)
		}
	})

	t.Run("古いバージョンはConflictで行は変わらない", func(t *testing.T) {
		stale := newBook(t, "9780000000033", "repo-test-stale", nil, 1)
		stale.ID, stale.Version = b.ID, 1
		if err := repo.Update(context.Background(), stale); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if stale.Version != 1 {
			t.Fatalf("Version = %d, want unchanged 1", stale.Version)
		}
		after, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatal(err)
		}
		if after.Bibliography.Title() != "repo-test-after" || after.Version != 2 {
			t.Fatalf("row changed on conflict: %+v", after)
		}
	})
}

func TestRepository_Delete(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)
	b := newBook(t, "9780000000040", "repo-test-delete", nil, 3)
	createBook(t, db, b)

	if err := repo.Delete(context.Background(), b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), b.ID); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("FindByID after delete: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(context.Background(), b.ID); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("Delete again: err = %v, want ErrNotFound", err)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := pgbook.NewRepository(connectTestDB(t))
	if _, err := repo.FindByID(context.Background(), domainbook.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestRepository_TitleOverrideAndSummary(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)

	t.Run("書名の上書きと一言まとめを保存して読み込める", func(t *testing.T) {
		b := newBook(t, "9780000001016", "カタログの書名", nil, 4)
		title, _ := domainbook.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		createBook(t, db, b)

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.TitleOverride == nil || got.TitleOverride.String() != "正しい書名" || got.Summary.String() != "一言まとめ" {
			t.Fatalf("got override=%v summary=%q", got.TitleOverride, got.Summary.String())
		}
	})

	t.Run("上書きを外して更新するとNULLになる", func(t *testing.T) {
		b := newBook(t, "9780000001023", "カタログの書名", nil, 4)
		title, _ := domainbook.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		createBook(t, db, b)

		b.OverrideTitle(nil)
		if err := repo.Update(context.Background(), b); err != nil {
			t.Fatalf("Update: %v", err)
		}
		var override *string
		db.Raw("SELECT title_override FROM book WHERE id = ?", int64(b.ID)).Scan(&override)
		if override != nil {
			t.Fatalf("title_override = %q, want NULL", *override)
		}
	})

	t.Run("一言まとめと上書きを変えて更新すると読み込みに反映される", func(t *testing.T) {
		b := newBook(t, "9780000002310", "カタログの書名", nil, 4)
		createBook(t, db, b)

		newSummary, _ := domainbook.NewSummary("読み返してのまとめ")
		b.ChangeReview(newSummary, b.Comment, b.Rating, b.Tags, b.Version)
		newTitle, _ := domainbook.NewTitle("新しい上書き")
		b.OverrideTitle(&newTitle)
		if err := repo.Update(context.Background(), b); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Summary.String() != "読み返してのまとめ" {
			t.Fatalf("Summary = %q, want 読み返してのまとめ", got.Summary.String())
		}
		if got.TitleOverride == nil || got.TitleOverride.String() != "新しい上書き" {
			t.Fatalf("TitleOverride = %v, want 新しい上書き", got.TitleOverride)
		}
	})
}

func TestRepository_Tags(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)

	// テスト用のタグを2つ作る
	tagA := mustCreateTag(t, db, "repo-test-タグA")
	tagB := mustCreateTag(t, db, "repo-test-タグB")
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'repo-test-%'") })

	t.Run("タグ付きで作成して読み込める", func(t *testing.T) {
		b := newBook(t, "9780000002419", "カタログの書名", nil, 4, domaintag.ID(tagA), domaintag.ID(tagB))
		createBook(t, db, b)

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		ids := got.Tags.IDs()
		if len(ids) != 2 {
			t.Fatalf("Tags = %v, want 2 tags", ids)
		}
	})

	t.Run("更新でタグを差し替えると入れ替わる", func(t *testing.T) {
		b := newBook(t, "9780000002426", "カタログの書名", nil, 4, domaintag.ID(tagA))
		createBook(t, db, b)

		newTags, err := domainbook.NewTagSelection([]domaintag.ID{domaintag.ID(tagB)})
		if err != nil {
			t.Fatal(err)
		}
		b.ChangeReview(b.Summary, b.Comment, b.Rating, newTags, b.Version)
		if err := repo.Update(context.Background(), b); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		ids := got.Tags.IDs()
		if len(ids) != 1 || ids[0] != domaintag.ID(tagB) {
			t.Fatalf("Tags = %v, want [%d]", ids, tagB)
		}
	})

	t.Run("タグを削除すると本から自動で外れる", func(t *testing.T) {
		soloTag := mustCreateTag(t, db, "repo-test-単独")
		b := newBook(t, "9780000002433", "カタログの書名", nil, 4, domaintag.ID(soloTag))
		createBook(t, db, b)

		db.Exec("DELETE FROM tag WHERE id = ?", soloTag)

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if len(got.Tags.IDs()) != 0 {
			t.Fatalf("Tags = %v, want empty after the tag was deleted", got.Tags.IDs())
		}
	})

	t.Run("タグを空にすると外れる", func(t *testing.T) {
		b := newBook(t, "9780000002464", "カタログの書名", nil, 4, domaintag.ID(tagA), domaintag.ID(tagB))
		createBook(t, db, b)

		empty, err := domainbook.NewTagSelection(nil)
		if err != nil {
			t.Fatal(err)
		}
		b.ChangeReview(b.Summary, b.Comment, b.Rating, empty, b.Version)
		if err := repo.Update(context.Background(), b); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if ids := got.Tags.IDs(); len(ids) != 0 {
			t.Fatalf("Tags = %v, want empty after clearing", ids)
		}
	})

	t.Run("タグの並び順はIDの昇順で安定している", func(t *testing.T) {
		// 付ける順を逆（tagB, tagA）にしても、読み込みはID昇順で返す。
		b := newBook(t, "9780000002488", "カタログの書名", nil, 4, domaintag.ID(tagB), domaintag.ID(tagA))
		createBook(t, db, b)

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		want := []domaintag.ID{domaintag.ID(tagA), domaintag.ID(tagB)}
		if ids := got.Tags.IDs(); len(ids) != 2 || ids[0] != want[0] || ids[1] != want[1] {
			t.Fatalf("Tags = %v, want %v (ID昇順)", ids, want)
		}
	})
}

func TestRepository_CreateAll(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)
	ctx := context.Background()
	t.Cleanup(func() {
		db.Exec("DELETE FROM book WHERE title LIKE 'seed-test-%'")
		db.Exec("DELETE FROM tag WHERE name LIKE 'seed-test-%'")
	})

	t.Run("全冊をタグごと保存し、IDとバージョンを設定する", func(t *testing.T) {
		tagID := mustCreateTag(t, db, "seed-test-タグ")
		books := []*domainbook.Book{
			newBook(t, "9780000003300", "seed-test-1冊目", nil, 3, domaintag.ID(tagID)),
			newBook(t, "9780000003317", "seed-test-2冊目", nil, 4),
		}
		if err := repo.CreateAll(ctx, books); err != nil {
			t.Fatalf("CreateAll: %v", err)
		}
		for _, b := range books {
			if b.ID == 0 || b.Version != 1 {
				t.Fatalf("%s: ID=%d Version=%d, want an assigned ID and version 1", b.ISBN.String(), b.ID, b.Version)
			}
		}
		got, err := repo.FindByID(ctx, books[0].ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if ids := got.Tags.IDs(); len(ids) != 1 || ids[0] != domaintag.ID(tagID) {
			t.Fatalf("Tags = %v, want [%d]", ids, tagID)
		}
	})

	t.Run("同じISBNがあればConflictを返し、1冊も残さない", func(t *testing.T) {
		createBook(t, db, newBook(t, "9780000003331", "seed-test-既存", nil, 3))
		books := []*domainbook.Book{
			newBook(t, "9780000003324", "seed-test-先に入る冊", nil, 3),
			newBook(t, "9780000003331", "seed-test-重複", nil, 3),
		}
		if err := repo.CreateAll(ctx, books); !errors.Is(err, domaincommon.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM book WHERE isbn = ?", "9780000003324").Scan(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("book 9780000003324 remains (%d rows), want the whole batch rolled back", count)
		}
		if books[0].ID != 0 || books[0].Version != 0 {
			t.Fatalf("rolled-back book has ID=%d Version=%d, want 0/0 (no row exists)", books[0].ID, books[0].Version)
		}
	})
}
