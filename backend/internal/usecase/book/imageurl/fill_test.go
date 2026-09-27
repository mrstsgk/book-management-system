package imageurl_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/imageurl"
)

// fakeImages is a hand-written Fake for book.ImageStorage; only URL is used here.
type fakeImages struct {
	url    string
	err    error
	gotKey *book.ImageKey
}

func (f *fakeImages) Put(context.Context, book.ImageKey, book.Image, io.Reader) error { return nil }
func (f *fakeImages) Delete(context.Context, book.ImageKey) error                     { return nil }

func (f *fakeImages) URL(_ context.Context, key book.ImageKey) (string, error) {
	f.gotKey = &key
	return f.url, f.err
}

func ptr(s string) *string { return &s }

func TestFill(t *testing.T) {
	t.Parallel()

	t.Run("画像キーからURLを詰める", func(t *testing.T) {
		t.Parallel()
		images := &fakeImages{url: "https://storage.example/signed"}
		d := &book.BookDetail{ImageKey: ptr("books/1/a.png")}

		if err := imageurl.Fill(context.Background(), images, d); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.ImageURL == nil || *d.ImageURL != "https://storage.example/signed" {
			t.Fatalf("ImageURL = %v", d.ImageURL)
		}
		if images.gotKey == nil || images.gotKey.String() != "books/1/a.png" {
			t.Fatalf("URL requested for %v, want books/1/a.png", images.gotKey)
		}
	})

	t.Run("画像がなければURLはnilでストレージを呼ばない", func(t *testing.T) {
		t.Parallel()
		images := &fakeImages{}
		d := &book.BookDetail{}

		if err := imageurl.Fill(context.Background(), images, d); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.ImageURL != nil || images.gotKey != nil {
			t.Fatalf("ImageURL = %v, storage called = %v; want neither", d.ImageURL, images.gotKey != nil)
		}
	})

	t.Run("ストレージのエラーはそのまま返りURLは詰めない", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("storage down")
		d := &book.BookDetail{ImageKey: ptr("books/1/a.png")}

		if err := imageurl.Fill(context.Background(), &fakeImages{err: wantErr}, d); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if d.ImageURL != nil {
			t.Fatalf("ImageURL = %v, want nil on failure", d.ImageURL)
		}
	})
}
