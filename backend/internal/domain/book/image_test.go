package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		contentType string
		size        int64
		wantExt     string
		wantErr     bool
	}{
		{name: "JPEGは有効", contentType: "image/jpeg", size: 1024, wantExt: ".jpg"},
		{name: "PNGは有効", contentType: "image/png", size: 1024, wantExt: ".png"},
		{name: "WebPは有効", contentType: "image/webp", size: 1024, wantExt: ".webp"},
		{name: "1バイトは有効", contentType: "image/png", size: 1, wantExt: ".png"},
		{name: "上限ちょうどは有効", contentType: "image/png", size: book.ImageMaxBytes, wantExt: ".png"},
		{name: "上限+1バイトはエラー", contentType: "image/png", size: book.ImageMaxBytes + 1, wantErr: true},
		{name: "0バイトはエラー", contentType: "image/png", size: 0, wantErr: true},
		{name: "GIFはエラー", contentType: "image/gif", size: 1024, wantErr: true},
		{name: "SVGはエラー", contentType: "image/svg+xml", size: 1024, wantErr: true},
		{name: "画像以外はエラー", contentType: "application/pdf", size: 1024, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewImage(tt.contentType, tt.size)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ContentType() != tt.contentType || got.Size() != tt.size || got.Extension() != tt.wantExt {
				t.Fatalf("got %s/%d/%s, want %s/%d/%s", got.ContentType(), got.Size(), got.Extension(), tt.contentType, tt.size, tt.wantExt)
			}
		})
	}
}

func TestNewImageKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "通常のキーは有効", in: "books/1/abc.png"},
		{name: "255バイトちょうどは有効", in: strings.Repeat("a", 255)},
		{name: "256バイトはエラー", in: strings.Repeat("a", 256), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "先頭スラッシュはエラー", in: "/books/1/abc.png", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewImageKey(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.String() != tt.in {
				t.Fatalf("got (%q, %v), want (%q, nil)", got.String(), err, tt.in)
			}
		})
	}
}
