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
)

// fakeCatalog is a hand-written Fake for domainbook.BookCatalog.
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
	if err != nil {
		t.Fatal(err)
	}
	return &c
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
		primary := &fakeCatalog{entry: &domainbook.CatalogEntry{Title: "openbd", Cover: openbdCover}}
		fallback := &fakeCatalog{}
		got, err := catalog.NewChain(primary, fallback).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != openbdCover || fallback.calls != 0 {
			t.Fatalf("got (%+v, %v), fallback calls=%d", got, err, fallback.calls)
		}
	})

	t.Run("primaryに書影が無ければfallbackの書影だけを補う", func(t *testing.T) {
		primary := &fakeCatalog{entry: &domainbook.CatalogEntry{Title: "openbd の書名"}}
		fallback := &fakeCatalog{entry: &domainbook.CatalogEntry{Title: "楽天の書名", Cover: rakutenCover}}
		got, err := catalog.NewChain(primary, fallback).Lookup(context.Background(), isbn(t))
		if err != nil || got.Title != "openbd の書名" || got.Cover != rakutenCover {
			t.Fatalf("got (%+v, %v), want primary's bibliography with fallback's cover", got, err)
		}
	})

	t.Run("primaryに該当が無ければfallbackの結果を返す", func(t *testing.T) {
		primary := &fakeCatalog{err: common.ErrNotFound}
		want := &domainbook.CatalogEntry{Title: "楽天の書名", Cover: rakutenCover}
		got, err := catalog.NewChain(primary, &fakeCatalog{entry: want}).Lookup(context.Background(), isbn(t))
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v), want fallback's entry", got, err)
		}
	})

	t.Run("どちらにも該当が無ければErrNotFound", func(t *testing.T) {
		chain := catalog.NewChain(&fakeCatalog{err: common.ErrNotFound}, &fakeCatalog{err: common.ErrNotFound})
		if _, err := chain.Lookup(context.Background(), isbn(t)); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("fallbackの障害は書影なしで返しエラーにしない", func(t *testing.T) {
		primary := &fakeCatalog{entry: &domainbook.CatalogEntry{Title: "openbd"}}
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
		primary := &fakeCatalog{entry: &domainbook.CatalogEntry{Title: "openbd"}}
		got, err := catalog.NewChain(primary, nil).Lookup(context.Background(), isbn(t))
		if err != nil || got.Title != "openbd" || got.Cover != nil {
			t.Fatalf("got (%+v, %v)", got, err)
		}
	})
}
