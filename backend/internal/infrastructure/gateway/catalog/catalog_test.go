package catalog_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/catalog"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/googlebooks"
)

// fakeBibliography は書誌を返す domainbook.BookCatalog（openBD）の手書き Fake（docs/rules/testing.md）。
type fakeBibliography struct {
	entry *domainbook.CatalogEntry
	err   error
}

func (f *fakeBibliography) Lookup(context.Context, domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	return f.entry, f.err
}

// fakeCovers は Google Books の手書き Fake。
type fakeCovers struct {
	cover   *googlebooks.Cover
	err     error
	calls   int
	gotISBN string
}

func (f *fakeCovers) FindCover(_ context.Context, isbn string) (*googlebooks.Cover, error) {
	f.calls++
	f.gotISBN = isbn
	return f.cover, f.err
}

func openBDEntry(t *testing.T, cover *domainbook.Cover) *domainbook.CatalogEntry {
	t.Helper()
	isbn, err := domainbook.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	bib, err := domainbook.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オライリー・ジャパン", "201907")
	if err != nil {
		t.Fatal(err)
	}
	return &domainbook.CatalogEntry{ISBN: isbn, Bibliography: bib, Cover: cover}
}

func openBDCover(t *testing.T) *domainbook.Cover {
	t.Helper()
	c, err := domainbook.NewCover("https://cover.openbd.jp/9784873118703.jpg", domainbook.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestCatalog_Lookup(t *testing.T) {
	silenceLog(t)
	googleCover := &googlebooks.Cover{
		ImageURL: "https://books.google.com/books/content?id=abc&printsec=frontcover&img=1&zoom=1",
		PageURL:  "https://books.google.co.jp/books?id=abc",
	}

	t.Run("Google Booksに書影があれば、書誌はopenBDのまま書影だけ差し替える", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, openBDCover(t))}
		covers := &fakeCovers{cover: googleCover}
		got, err := catalog.New(bib, covers).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil {
			t.Fatal(err)
		}
		if got.Bibliography.Title() != "データ指向アプリケーションデザイン" {
			t.Fatalf("title = %q, want openBD's", got.Bibliography.Title())
		}
		if got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceGoogleBooks ||
			got.Cover.URL() != googleCover.ImageURL || got.Cover.PageURL() != googleCover.PageURL {
			t.Fatalf("cover = %+v, want the Google Books cover", got.Cover)
		}
		if covers.gotISBN != "9784873118703" {
			t.Fatalf("asked Google Books for %q, want the ISBN-13", covers.gotISBN)
		}
	})

	t.Run("Google Booksに書影が無ければopenBDの書影のまま", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, openBDCover(t))}
		got, err := catalog.New(bib, &fakeCovers{}).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil || got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceOpenBD {
			t.Fatalf("got (%+v, %v), want the openBD cover", got, err)
		}
	})

	t.Run("どちらにも書影が無ければ書影なし", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, nil)}
		got, err := catalog.New(bib, &fakeCovers{}).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil || got.Cover != nil {
			t.Fatalf("got (%+v, %v), want no cover", got, err)
		}
	})

	t.Run("Google Booksの失敗では止めず、openBDの結果を返す", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, openBDCover(t))}
		covers := &fakeCovers{err: errors.New("googlebooks: unexpected status 429")}
		got, err := catalog.New(bib, covers).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil || got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceOpenBD {
			t.Fatalf("got (%+v, %v), want the openBD result", got, err)
		}
	})

	t.Run("Google Booksの書影が不正なURLならopenBDの書影のまま", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, openBDCover(t))}
		covers := &fakeCovers{cover: &googlebooks.Cover{ImageURL: googleCover.ImageURL, PageURL: "not a url"}}
		got, err := catalog.New(bib, covers).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil || got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceOpenBD {
			t.Fatalf("got (%+v, %v), want the openBD cover", got, err)
		}
	})

	t.Run("openBDに無い本はGoogle Booksに問い合わせずNotFound", func(t *testing.T) {
		bib := &fakeBibliography{err: common.ErrNotFound}
		covers := &fakeCovers{cover: googleCover}
		isbn, _ := domainbook.NewISBN("9784873118703")
		_, err := catalog.New(bib, covers).Lookup(context.Background(), isbn)
		if !errors.Is(err, common.ErrNotFound) || covers.calls != 0 {
			t.Fatalf("err = %v (Google Books calls: %d), want NotFound without asking Google Books", err, covers.calls)
		}
	})

	t.Run("Google Booksが無ければ（キー未設定）openBDの結果をそのまま返す", func(t *testing.T) {
		bib := &fakeBibliography{entry: openBDEntry(t, openBDCover(t))}
		got, err := catalog.New(bib, nil).Lookup(context.Background(), bib.entry.ISBN)
		if err != nil || got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceOpenBD {
			t.Fatalf("got (%+v, %v), want the openBD result", got, err)
		}
	})
}
