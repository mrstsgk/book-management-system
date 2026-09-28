package catalog_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/catalog"
)

// fakeCatalog は domainbook.BookCatalog の手書き Fake（docs/rules/testing.md）。
type fakeCatalog struct {
	entry *domainbook.CatalogEntry
	err   error
	calls int
}

func (f *fakeCatalog) Lookup(context.Context, domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	f.calls++
	return f.entry, f.err
}

func cover(t *testing.T, url string, src domainbook.CoverSource) *domainbook.Cover {
	t.Helper()
	c, err := domainbook.NewCover(url, src)
	if src == domainbook.CoverSourceRakuten {
		c, err = domainbook.NewRakutenCover(url, "https://books.rakuten.co.jp/rb/1/", time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC))
	}
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func entry(t *testing.T, title string, c *domainbook.Cover) *domainbook.CatalogEntry {
	t.Helper()
	bib, err := domainbook.NewBibliography(title, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	return &domainbook.CatalogEntry{Bibliography: bib, Cover: c}
}

func isbn(t *testing.T) domainbook.ISBN {
	t.Helper()
	v, err := domainbook.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestChain_Lookup(t *testing.T) {
	silenceLog(t)
	openbdCover := cover(t, "https://cover.openbd.jp/1.jpg", domainbook.CoverSourceOpenBD)
	rakutenCover := cover(t, "https://thumbnail.image.rakuten.co.jp/1.jpg", domainbook.CoverSourceRakuten)

	t.Run("primaryに書影があればfallbackを呼ばない", func(t *testing.T) {
		primary := &fakeCatalog{entry: entry(t, "openbd", openbdCover)}
		fallback := &fakeCatalog{}
		got, err := catalog.NewChain(primary, fallback).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != openbdCover || fallback.calls != 0 {
			t.Fatalf("got (%+v, %v), fallback calls=%d", got, err, fallback.calls)
		}
	})

	t.Run("primaryに書影が無ければfallbackの書影だけを補う", func(t *testing.T) {
		primary := &fakeCatalog{entry: entry(t, "openbd の書名", nil)}
		fallback := &fakeCatalog{entry: entry(t, "楽天の書名", rakutenCover)}
		got, err := catalog.NewChain(primary, fallback).Lookup(context.Background(), isbn(t))
		if err != nil || got.Bibliography.Title() != "openbd の書名" || got.Cover != rakutenCover {
			t.Fatalf("got (%+v, %v), want primary's bibliography with fallback's cover", got, err)
		}
	})

	t.Run("primaryに該当が無ければfallbackを呼ばずにErrNotFound", func(t *testing.T) {
		fallback := &fakeCatalog{entry: entry(t, "楽天の書名", rakutenCover)}
		chain := catalog.NewChain(&fakeCatalog{err: common.ErrNotFound}, fallback)
		if _, err := chain.Lookup(context.Background(), isbn(t)); !errors.Is(err, common.ErrNotFound) || fallback.calls != 0 {
			t.Fatalf("err = %v, fallback calls=%d, want ErrNotFound without calling fallback", err, fallback.calls)
		}
	})

	t.Run("fallbackにも書影が無ければ書影なしで返す", func(t *testing.T) {
		primary := &fakeCatalog{entry: entry(t, "openbd の書名", nil)}
		got, err := catalog.NewChain(primary, &fakeCatalog{err: common.ErrNotFound}).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != nil {
			t.Fatalf("got (%+v, %v), want primary's entry without cover", got, err)
		}
	})

	t.Run("fallbackの障害は書影なしで返しエラーにしない", func(t *testing.T) {
		primary := &fakeCatalog{entry: entry(t, "openbd", nil)}
		got, err := catalog.NewChain(primary, &fakeCatalog{err: errors.New("rakuten: timeout")}).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != nil {
			t.Fatalf("got (%+v, %v), want the primary entry without a cover", got, err)
		}
	})

	t.Run("primaryの障害はfallbackを呼ばずにエラーを返す", func(t *testing.T) {
		wantErr := errors.New("openbd: timeout")
		fallback := &fakeCatalog{}
		_, err := catalog.NewChain(&fakeCatalog{err: wantErr}, fallback).Lookup(context.Background(), isbn(t))
		if !errors.Is(err, wantErr) || fallback.calls != 0 {
			t.Fatalf("err = %v, fallback calls=%d; want the primary error and no fallback", err, fallback.calls)
		}
	})

	t.Run("fallbackがnilならprimaryだけを使う", func(t *testing.T) {
		primary := &fakeCatalog{entry: entry(t, "openbd", nil)}
		got, err := catalog.NewChain(primary, nil).Lookup(context.Background(), isbn(t))
		if err != nil || got.Bibliography.Title() != "openbd" || got.Cover != nil {
			t.Fatalf("got (%+v, %v)", got, err)
		}
	})
}
