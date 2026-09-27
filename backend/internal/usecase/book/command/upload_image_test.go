package command_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

func uploadCmd() command.UploadImageCommand {
	return command.UploadImageCommand{BookID: 10, ContentType: "image/png", Size: 3, Body: bytes.NewReader([]byte("png"))}
}

func fixedToken() string { return "tok" }

func TestUploadImageUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("画像を保存して書籍のキーを差し替え詳細を返す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t, book.Unpublished)}
		images := &fakeImages{url: "https://storage.example/"}
		key := "books/10/tok.png"
		uc := &command.UploadImageUsecaseImpl{Books: books, Images: images, Details: &fakeDetails{detail: &book.BookDetail{ID: 10, ImageKey: &key}}, NewToken: fixedToken}

		got, err := uc.Execute(context.Background(), uploadCmd())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images.put) != 1 || images.put[0].String() != key || string(images.putBody) != "png" {
			t.Fatalf("put %v with %q, want %s with the uploaded bytes", images.put, images.putBody, key)
		}
		if books.updated == nil || books.updated.ImageKey == nil || books.updated.ImageKey.String() != key {
			t.Fatalf("book not updated with the new key: %+v", books.updated)
		}
		if len(images.deleted) != 0 {
			t.Fatalf("deleted %v, want nothing for a book without a previous image", images.deleted)
		}
		if got.ImageURL == nil || *got.ImageURL != "https://storage.example/"+key {
			t.Fatalf("ImageURL = %v", got.ImageURL)
		}
	})

	t.Run("差し替えに成功したら古い画像を消す", func(t *testing.T) {
		t.Parallel()
		existing := existingBook(t, book.Unpublished)
		old, err := book.NewImageKey("books/10/old.png")
		if err != nil {
			t.Fatal(err)
		}
		existing.ReplaceImage(old)
		images := &fakeImages{}
		uc := &command.UploadImageUsecaseImpl{Books: &fakeBooks{findByID: existing}, Images: images, Details: &fakeDetails{detail: &book.BookDetail{}}, NewToken: fixedToken}

		if _, err := uc.Execute(context.Background(), uploadCmd()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images.deleted) != 1 || images.deleted[0] != old {
			t.Fatalf("deleted %v, want only the old image", images.deleted)
		}
	})

	t.Run("古い画像の削除に失敗しても成功として返す", func(t *testing.T) {
		t.Parallel()
		existing := existingBook(t, book.Unpublished)
		old, _ := book.NewImageKey("books/10/old.png")
		existing.ReplaceImage(old)
		uc := &command.UploadImageUsecaseImpl{
			Books: &fakeBooks{findByID: existing}, Images: &fakeImages{delErr: errors.New("storage down")},
			Details: &fakeDetails{detail: &book.BookDetail{}}, NewToken: fixedToken,
		}

		if _, err := uc.Execute(context.Background(), uploadCmd()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("書籍の保存に失敗したら新しい画像を消してエラーを返す", func(t *testing.T) {
		t.Parallel()
		images := &fakeImages{}
		uc := &command.UploadImageUsecaseImpl{
			Books: &fakeBooks{findByID: existingBook(t, book.Unpublished), updateErr: common.ErrConflict}, Images: images,
			Details: &fakeDetails{}, NewToken: fixedToken,
		}

		if _, err := uc.Execute(context.Background(), uploadCmd()); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if len(images.deleted) != 1 || images.deleted[0].String() != "books/10/tok.png" {
			t.Fatalf("deleted %v, want the just-uploaded image", images.deleted)
		}
	})

	invalid := []struct {
		name   string
		mutate func(*command.UploadImageCommand)
	}{
		{name: "画像以外はエラーで保存しない", mutate: func(c *command.UploadImageCommand) { c.ContentType = "text/plain" }},
		{name: "サイズ超過はエラーで保存しない", mutate: func(c *command.UploadImageCommand) { c.Size = book.ImageMaxBytes + 1 }},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			books := &fakeBooks{findByID: existingBook(t, book.Unpublished)}
			images := &fakeImages{}
			uc := &command.UploadImageUsecaseImpl{Books: books, Images: images, Details: &fakeDetails{}, NewToken: fixedToken}
			cmd := uploadCmd()
			tt.mutate(&cmd)

			if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if len(images.put) != 0 || books.updated != nil {
				t.Fatal("neither storage nor repository may be written for an invalid image")
			}
		})
	}

	t.Run("存在しない書籍はNotFoundで画像を保存しない", func(t *testing.T) {
		t.Parallel()
		images := &fakeImages{}
		uc := &command.UploadImageUsecaseImpl{Books: &fakeBooks{findByIDErr: common.ErrNotFound}, Images: images, Details: &fakeDetails{}}

		if _, err := uc.Execute(context.Background(), uploadCmd()); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if len(images.put) != 0 {
			t.Fatal("storage must not be written for an unknown book")
		}
	})

	t.Run("ストレージへの保存に失敗したら書籍は更新しない", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("storage down")
		books := &fakeBooks{findByID: existingBook(t, book.Unpublished)}
		uc := &command.UploadImageUsecaseImpl{Books: books, Images: &fakeImages{putErr: wantErr}, Details: &fakeDetails{}, NewToken: fixedToken}

		if _, err := uc.Execute(context.Background(), uploadCmd()); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if books.updated != nil {
			t.Fatal("the book must not point at an image that was never stored")
		}
	})

	t.Run("トークン未指定でもランダムなキーで保存する", func(t *testing.T) {
		t.Parallel()
		images := &fakeImages{}
		uc := &command.UploadImageUsecaseImpl{Books: &fakeBooks{findByID: existingBook(t, book.Unpublished)}, Images: images, Details: &fakeDetails{detail: &book.BookDetail{}}}

		if _, err := uc.Execute(context.Background(), uploadCmd()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := uc.Execute(context.Background(), uploadCmd()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images.put) != 2 || images.put[0] == images.put[1] {
			t.Fatalf("keys %v, want two distinct keys", images.put)
		}
	})
}
