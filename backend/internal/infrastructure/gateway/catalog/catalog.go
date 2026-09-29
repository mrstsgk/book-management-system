// Package catalog は外部カタログの組み合わせ方（どこから書誌を取り、どの順で書影を探すか）を1か所に書く。
// 書誌: openBD。書影: Google Books → openBD → なし（docs/superpowers/specs/2026-09-29-cover-sources-design.md）。
package catalog

import (
	"context"
	"log/slog"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/googlebooks"
)

// coverFinder は書影だけを返す提供元（Google Books）。
type coverFinder interface {
	FindCover(ctx context.Context, isbn string) (*googlebooks.Cover, error)
}

type catalog struct {
	openbd domainbook.BookCatalog
	google coverFinder
}

// New は google が nil（API キー未設定）なら openBD だけを返す。
func New(openbd domainbook.BookCatalog, google coverFinder) domainbook.BookCatalog {
	if google == nil {
		return openbd
	}
	return &catalog{openbd: openbd, google: google}
}

// Lookup は書誌と openBD の書影を openBD から取り、Google Books に書影があれば差し替える。
// Google Books の失敗では止めない（書影が openBD のもの、または無くなるだけで、登録・更新は続けられるため）。
func (c *catalog) Lookup(ctx context.Context, isbn domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	entry, err := c.openbd.Lookup(ctx, isbn)
	if err != nil {
		return nil, err
	}
	found, err := c.google.FindCover(ctx, isbn.String())
	if err != nil {
		slog.WarnContext(ctx, "google books cover lookup failed; keeping the openBD result", "isbn", isbn.String(), "error", err)
		return entry, nil
	}
	if found == nil {
		return entry, nil
	}
	cover, err := domainbook.NewGoogleBooksCover(found.ImageURL, found.PageURL)
	if err != nil {
		slog.WarnContext(ctx, "google books returned an unusable cover; keeping the openBD result", "isbn", isbn.String(), "error", err)
		return entry, nil
	}
	entry.Cover = &cover
	return entry, nil
}
