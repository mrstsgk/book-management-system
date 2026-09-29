package command_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

// isbnCatalog は ISBN ごとに結果を変えられる book.BookCatalog の手書き Fake。
type isbnCatalog struct {
	entries map[string]*book.CatalogEntry
	errs    map[string]error
	asked   []string
	// onLookup は問い合わせのたびに呼ぶ（途中でキャンセルされたときの確認用）。
	onLookup func()
}

func (f *isbnCatalog) Lookup(_ context.Context, isbn book.ISBN) (*book.CatalogEntry, error) {
	f.asked = append(f.asked, isbn.String())
	if f.onLookup != nil {
		f.onLookup()
	}
	if err := f.errs[isbn.String()]; err != nil {
		return nil, err
	}
	if e, ok := f.entries[isbn.String()]; ok {
		return e, nil
	}
	return nil, common.ErrNotFound
}

func coverlessBook(t *testing.T, isbn string, id book.ID) *book.Book {
	t.Helper()
	i, err := book.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, _ := book.NewBibliography("書影の無い本", "", "", "")
	summary, _ := book.NewSummary("まとめ")
	comment, _ := book.NewComment("感想")
	rating, _ := book.NewRating(3)
	b := book.New(i, bib, nil, summary, comment, rating, book.TagSelection{})
	b.ID, b.Version = id, 1
	return b
}

func entryWithCover(t *testing.T, isbn, title string, cover *book.Cover) *book.CatalogEntry {
	t.Helper()
	i, _ := book.NewISBN(isbn)
	bib, err := book.NewBibliography(title, "著者", "出版社", "202401")
	if err != nil {
		t.Fatal(err)
	}
	return &book.CatalogEntry{ISBN: i, Bibliography: bib, Cover: cover}
}

func googleCover(t *testing.T) *book.Cover {
	t.Helper()
	c, err := book.NewGoogleBooksCover("https://books.google.com/books/content?id=a&img=1", "https://books.google.co.jp/books?id=a")
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func quietLogs(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

const (
	isbnA = "9784873118703"
	isbnB = "9784295016090"
	isbnC = "9784297146221"
)

func TestFillMissingCoversUsecase_Execute(t *testing.T) {
	quietLogs(t)

	t.Run("書影が見つかった本は、取り直した書誌と書影で保存する", func(t *testing.T) {
		books := &fakeBooks{coverless: []*book.Book{coverlessBook(t, isbnA, 1)}}
		cover := googleCover(t)
		catalog := &isbnCatalog{entries: map[string]*book.CatalogEntry{isbnA: entryWithCover(t, isbnA, "取り直した書名", cover)}}
		if err := (&command.FillMissingCoversUsecaseImpl{Books: books, Catalog: catalog}).Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(books.updatedAll) != 1 {
			t.Fatalf("saved %d books, want 1", len(books.updatedAll))
		}
		got := books.updatedAll[0]
		if got.Cover == nil || *got.Cover != *cover || got.Bibliography.Title() != "取り直した書名" {
			t.Fatalf("saved (cover=%+v, title=%q), want the found cover and bibliography", got.Cover, got.Bibliography.Title())
		}
	})

	t.Run("書影が見つからない本は保存しない（版を上げない）", func(t *testing.T) {
		books := &fakeBooks{coverless: []*book.Book{coverlessBook(t, isbnA, 1)}}
		catalog := &isbnCatalog{entries: map[string]*book.CatalogEntry{isbnA: entryWithCover(t, isbnA, "書名", nil)}}
		if err := (&command.FillMissingCoversUsecaseImpl{Books: books, Catalog: catalog}).Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(books.updatedAll) != 0 {
			t.Fatalf("saved %d books, want none", len(books.updatedAll))
		}
	})

	t.Run("カタログや保存の失敗があっても、残りの本は埋める", func(t *testing.T) {
		books := &fakeBooks{
			coverless:  []*book.Book{coverlessBook(t, isbnA, 1), coverlessBook(t, isbnB, 2), coverlessBook(t, isbnC, 3)},
			updateErrs: map[string]error{isbnB: common.ErrConflict},
		}
		catalog := &isbnCatalog{
			entries: map[string]*book.CatalogEntry{
				isbnB: entryWithCover(t, isbnB, "B", googleCover(t)),
				isbnC: entryWithCover(t, isbnC, "C", googleCover(t)),
			},
			errs: map[string]error{isbnA: errors.New("openbd: timeout")},
		}
		if err := (&command.FillMissingCoversUsecaseImpl{Books: books, Catalog: catalog}).Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(catalog.asked) != 3 {
			t.Fatalf("asked the catalog for %v, want all 3 books", catalog.asked)
		}
		if len(books.updatedAll) != 1 || books.updatedAll[0].ISBN.String() != isbnC {
			t.Fatalf("saved %d books, want only %s", len(books.updatedAll), isbnC)
		}
	})

	t.Run("書影の無い本を探せなければエラーを返す", func(t *testing.T) {
		want := errors.New("db: down")
		books := &fakeBooks{coverlessErr: want}
		err := (&command.FillMissingCoversUsecaseImpl{Books: books, Catalog: &isbnCatalog{}}).Execute(context.Background())
		if !errors.Is(err, want) {
			t.Fatalf("err = %v, want %v", err, want)
		}
	})

	t.Run("キャンセルされたら残りの本に問い合わせず、エラーにしない", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		books := &fakeBooks{coverless: []*book.Book{coverlessBook(t, isbnA, 1), coverlessBook(t, isbnB, 2)}}
		catalog := &isbnCatalog{onLookup: cancel}
		if err := (&command.FillMissingCoversUsecaseImpl{Books: books, Catalog: catalog}).Execute(ctx); err != nil {
			t.Fatalf("err = %v, want nil (shutdown is not a failure)", err)
		}
		if len(catalog.asked) != 1 {
			t.Fatalf("asked the catalog for %v, want only the first book", catalog.asked)
		}
	})
}
