package command

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// UploadImageCommand is the input contract for UploadImageUsecase (the boundary
// crossed from Presentation). It carries data only — no logic.
type UploadImageCommand struct {
	BookID      int64
	ContentType string
	Size        int64
	Body        io.Reader
}

type UploadImageUsecase interface {
	Execute(ctx context.Context, cmd UploadImageCommand) (*book.BookDetail, error)
}

type UploadImageUsecaseImpl struct {
	Books   book.Repository
	Images  book.ImageStorage
	Details book.Query
	// NewToken makes the random part of the object key; nil uses crypto/rand.
	NewToken func() string
}

// Execute は画像をストレージへ保存してから書籍の画像キーを差し替える。書籍の保存に失敗したら新しい画像を消し、成功したら古い画像を消す。
func (u *UploadImageUsecaseImpl) Execute(ctx context.Context, cmd UploadImageCommand) (*book.BookDetail, error) {
	b, err := u.Books.FindByID(ctx, book.ID(cmd.BookID))
	if err != nil {
		return nil, err
	}
	image, err := book.NewImage(cmd.ContentType, cmd.Size)
	if err != nil {
		return nil, err
	}
	// A fresh key per upload (never overwriting) keeps already-issued URLs and caches from serving a different image.
	key, err := book.NewImageKey(fmt.Sprintf("books/%d/%s%s", b.ID, u.token(), image.Extension()))
	if err != nil {
		return nil, err
	}
	if err := u.Images.Put(ctx, key, image, cmd.Body); err != nil {
		return nil, err
	}
	previous := b.ReplaceImage(key)
	if err := u.Books.Update(ctx, b); err != nil {
		u.deleteQuietly(ctx, key)
		return nil, err
	}
	if previous != nil {
		u.deleteQuietly(ctx, *previous)
	}
	return detailOf(ctx, u.Details, u.Images, b.ID)
}

// deleteQuietly only logs: the book row is already consistent, and an orphaned object costs storage, not correctness.
func (u *UploadImageUsecaseImpl) deleteQuietly(ctx context.Context, key book.ImageKey) {
	if err := u.Images.Delete(ctx, key); err != nil {
		slog.WarnContext(ctx, "failed to delete image object", "key", key.String(), "error", err)
	}
}

func (u *UploadImageUsecaseImpl) token() string {
	if u.NewToken != nil {
		return u.NewToken()
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error (Go 1.24+).
	return hex.EncodeToString(b)
}
