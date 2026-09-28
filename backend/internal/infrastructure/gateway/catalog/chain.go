// Package catalog combines book catalogs into the single domainbook.BookCatalog the use cases see.
package catalog

import (
	"context"
	"errors"
	"log/slog"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// chain asks primary (openBD) first and uses fallback (Rakuten) only for what primary
// lacks: the whole entry when primary doesn't know the ISBN, otherwise just the cover.
type chain struct {
	primary  domainbook.BookCatalog
	fallback domainbook.BookCatalog
}

// NewChain returns primary alone when fallback is nil (e.g. no Rakuten keys configured).
func NewChain(primary, fallback domainbook.BookCatalog) domainbook.BookCatalog {
	if fallback == nil {
		return primary
	}
	return &chain{primary: primary, fallback: fallback}
}

// Lookup は primary の結果を優先し、該当なしなら fallback の結果を、書影だけ無ければ fallback の書影を補う。fallback の失敗は書影の補完を諦めるだけにする。
func (c *chain) Lookup(ctx context.Context, isbn domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	entry, err := c.primary.Lookup(ctx, isbn)
	if errors.Is(err, common.ErrNotFound) {
		return c.fallback.Lookup(ctx, isbn)
	}
	if err != nil {
		return nil, err
	}
	if entry.Cover != nil {
		return entry, nil
	}
	extra, err := c.fallback.Lookup(ctx, isbn)
	switch {
	case err == nil:
		entry.Cover = extra.Cover
	case !errors.Is(err, common.ErrNotFound):
		slog.WarnContext(ctx, "cover fallback lookup failed", "isbn", isbn.String(), "error", err)
	}
	return entry, nil
}
