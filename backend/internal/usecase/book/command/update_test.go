package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

func existingBook(t *testing.T) *book.Book {
	t.Helper()
	entry := catalogEntry(t, "旧い書名", mustCover(t, "https://cover.openbd.jp/old.jpg"))
	summary, _ := book.NewSummary("最初のまとめ")
	comment, _ := book.NewComment("最初の感想")
	rating, _ := book.NewRating(3)
	b := book.New(entry.ISBN, entry.Bibliography, entry.Cover, summary, comment, rating, book.TagSelection{})
	title, _ := book.NewTitle("前の上書き")
	b.OverrideTitle(&title)
	b.ID, b.Version = 10, 2
	return b
}

func TestUpdateUsecase_Execute(t *testing.T) {
	t.Parallel()
	cmd := command.UpdateCommand{ID: 10, Summary: "読み返したまとめ", TitleOverride: "新しい上書き", TagIDs: []int64{3}, Comment: "読み返した", Rating: 5, Version: 2}

	t.Run("感想・評価を差し替え、書誌と書影を取り直して保存する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t)}
		details := &fakeDetails{detail: &book.BookDetail{ID: 10}}
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "新しい書名", nil)}, Details: details, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		u := books.updated
		if u == nil || u.Summary.String() != "読み返したまとめ" || u.Comment.String() != "読み返した" || u.Rating.Int() != 5 ||
			u.Bibliography.Title() != "新しい書名" || u.Cover != nil {
			t.Fatalf("updated %+v", u)
		}
		if u.TitleOverride == nil || u.TitleOverride.String() != "新しい上書き" {
			t.Fatalf("TitleOverride = %v, want it kept even though the catalog returned a new title", u.TitleOverride)
		}
		if ids := u.Tags.IDs(); len(ids) != 1 || ids[0] != 3 {
			t.Fatalf("Tags = %v, want [3]", ids)
		}
		if details.gotID != 10 {
			t.Fatalf("detail looked up for %d, want 10", details.gotID)
		}
	})

	t.Run("書名の上書きを空で送ると上書きを外す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t)}
		c := cmd
		c.TitleOverride = ""
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "新しい書名", nil)}, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.updated == nil || books.updated.TitleOverride != nil {
			t.Fatalf("updated %+v, want TitleOverride = nil", books.updated)
		}
	})

	t.Run("空の一言まとめはカタログも保存も呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t)}
		catalog := &fakeCatalog{}
		c := cmd
		c.Summary = ""
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if catalog.called != 0 || books.updated != nil {
			t.Fatal("neither the catalog nor the repository may be called for invalid input")
		}
	})

	keeps := []struct {
		name    string
		catalog *fakeCatalog
	}{
		{name: "カタログの障害では今の書誌と書影のまま感想を更新する", catalog: &fakeCatalog{err: errors.New("openbd: timeout")}},
		{name: "カタログに無くなっても今の書誌と書影のまま感想を更新する", catalog: &fakeCatalog{err: common.ErrNotFound}},
	}
	for _, tt := range keeps {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			books := &fakeBooks{findByID: existingBook(t)}
			uc := &command.UpdateUsecaseImpl{Books: books, Catalog: tt.catalog, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

			if _, err := uc.Execute(context.Background(), cmd); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			u := books.updated
			if u == nil || u.Comment.String() != "読み返した" || u.Bibliography.Title() != "旧い書名" || u.Cover == nil {
				t.Fatalf("updated %+v, want the new review with the current catalog data", u)
			}
		})
	}

	t.Run("不正な感想はエラーで保存せず、カタログも呼ばない", func(t *testing.T) {
		t.Parallel()
		existing := existingBook(t)
		books := &fakeBooks{findByID: existing}
		catalog := &fakeCatalog{}
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}
		c := cmd
		c.Comment = ""

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.updated != nil || catalog.called != 0 || existing.Comment.String() != "最初の感想" {
			t.Fatal("nothing may change when the input is invalid")
		}
	})

	t.Run("存在しない本はNotFound", func(t *testing.T) {
		t.Parallel()
		uc := &command.UpdateUsecaseImpl{Books: &fakeBooks{findByIDErr: common.ErrNotFound}, Catalog: &fakeCatalog{}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("楽観的ロックの競合はConflictを返す", func(t *testing.T) {
		t.Parallel()
		uc := &command.UpdateUsecaseImpl{Books: &fakeBooks{findByID: existingBook(t), updateErr: common.ErrConflict}, Catalog: &fakeCatalog{entry: catalogEntry(t, "x", nil)}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})

	t.Run("存在しないタグIDはカタログも保存も呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t)}
		catalog := &fakeCatalog{}
		c := cmd
		c.TagIDs = []int64{999}
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: false}}

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if catalog.called != 0 || books.updated != nil {
			t.Fatal("neither the catalog nor the repository may be called for invalid input")
		}
	})

	t.Run("タグを11個指定するとカタログも保存も呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t)}
		catalog := &fakeCatalog{}
		c := cmd
		c.TagIDs = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
		uc := &command.UpdateUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if catalog.called != 0 || books.updated != nil {
			t.Fatal("neither the catalog nor the repository may be called for invalid input")
		}
	})
}
