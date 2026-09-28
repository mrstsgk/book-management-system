// Package catalog は複数の書籍カタログを、UseCase から見える1つの domainbook.BookCatalog にまとめる。
package catalog

import (
	"context"
	"errors"
	"log/slog"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// chain は primary（openBD）を優先し、primary に足りないものだけを fallback（楽天）で補う。
// primary が ISBN を知らなければ fallback の結果全体を、知っていれば書影だけを補う。
type chain struct {
	primary  domainbook.BookCatalog
	fallback domainbook.BookCatalog
}

// NewChain は fallback が nil（楽天のキーが未設定など）なら primary だけを返す。
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
