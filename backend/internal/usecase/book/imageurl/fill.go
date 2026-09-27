// Package imageurl turns a book's stored image key into a URL a client can fetch,
// shared by the command (create/update/upload) and query (get) use cases.
package imageurl

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// Fill sets d.ImageURL from d.ImageKey; a book without an image keeps a nil URL.
func Fill(ctx context.Context, images book.ImageStorage, d *book.BookDetail) error {
	if d.ImageKey == nil {
		return nil
	}
	key, err := book.NewImageKey(*d.ImageKey)
	if err != nil {
		return err
	}
	u, err := images.URL(ctx, key)
	if err != nil {
		return err
	}
	d.ImageURL = &u
	return nil
}
