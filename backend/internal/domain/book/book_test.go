package book_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func mustTitle(t *testing.T, s string) book.Title {
	t.Helper()
	v, err := book.NewTitle(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustPrice(t *testing.T, n int64) book.Price {
	t.Helper()
	v, err := book.NewPrice(n)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		authorIDs []author.ID
		wantErr   bool
	}{
		{name: "著者1人は有効", authorIDs: []author.ID{1}},
		{name: "著者複数は有効", authorIDs: []author.ID{1, 2}},
		{name: "著者0人はエラー", authorIDs: nil, wantErr: true},
		{name: "著者IDの重複はエラー", authorIDs: []author.ID{1, 2, 1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b, err := book.New(mustTitle(t, "人間失格"), mustPrice(t, 1500), tt.authorIDs, book.Unpublished)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.ID != 0 || b.Version != 0 || !reflect.DeepEqual(b.AuthorIDs, tt.authorIDs) {
				t.Fatalf("got %+v", b)
			}
		})
	}
}

func TestBook_Change(t *testing.T) {
	t.Parallel()
	newBook := func(t *testing.T, status book.PublishStatus) *book.Book {
		t.Helper()
		b, err := book.New(mustTitle(t, "人間失格"), mustPrice(t, 1500), []author.ID{1}, status)
		if err != nil {
			t.Fatal(err)
		}
		b.ID, b.Version = 10, 3
		return b
	}

	t.Run("未出版から出版済みへ内容を差し替えられる", func(t *testing.T) {
		t.Parallel()
		b := newBook(t, book.Unpublished)

		err := b.Change(mustTitle(t, "斜陽"), mustPrice(t, 800), []author.ID{2}, book.Published, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.Title.String() != "斜陽" || b.Price.Int64() != 800 || b.Status != book.Published ||
			!reflect.DeepEqual(b.AuthorIDs, []author.ID{2}) || b.ID != 10 || b.Version != 3 {
			t.Fatalf("got %+v", b)
		}
	})

	failures := []struct {
		name      string
		from      book.PublishStatus
		to        book.PublishStatus
		authorIDs []author.ID
	}{
		{name: "出版済みから未出版はエラーで状態は変わらない", from: book.Published, to: book.Unpublished, authorIDs: []author.ID{2}},
		{name: "著者0人はエラーで状態は変わらない", from: book.Unpublished, to: book.Unpublished, authorIDs: nil},
		{name: "著者IDの重複はエラーで状態は変わらない", from: book.Unpublished, to: book.Unpublished, authorIDs: []author.ID{2, 2}},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := newBook(t, tt.from)
			before := *b

			err := b.Change(mustTitle(t, "斜陽"), mustPrice(t, 800), tt.authorIDs, tt.to, 3)
			if !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if !reflect.DeepEqual(*b, before) {
				t.Fatalf("book changed on failure: got %+v, want %+v", *b, before)
			}
		})
	}
}

func TestBook_ChangeAmazonURL(t *testing.T) {
	t.Parallel()
	b, err := book.New(mustTitle(t, "人間失格"), mustPrice(t, 1500), []author.ID{1}, book.Unpublished)
	if err != nil {
		t.Fatal(err)
	}
	u, err := book.NewAmazonURL("https://www.amazon.co.jp/dp/4101006059")
	if err != nil {
		t.Fatal(err)
	}

	b.ChangeAmazonURL(&u)
	if b.AmazonURL == nil || *b.AmazonURL != u {
		t.Fatalf("AmazonURL = %v, want %v", b.AmazonURL, u)
	}

	b.ChangeAmazonURL(nil)
	if b.AmazonURL != nil {
		t.Fatalf("AmazonURL = %v, want nil after clearing", b.AmazonURL)
	}
}

func TestBook_ChangeAmazonURL_IsNotAffectedByLaterChangesToTheArgument(t *testing.T) {
	t.Parallel()
	b, err := book.New(mustTitle(t, "人間失格"), mustPrice(t, 1500), []author.ID{1}, book.Unpublished)
	if err != nil {
		t.Fatal(err)
	}
	u, err := book.NewAmazonURL("https://www.amazon.co.jp/dp/first")
	if err != nil {
		t.Fatal(err)
	}
	other, err := book.NewAmazonURL("https://www.amazon.co.jp/dp/other")
	if err != nil {
		t.Fatal(err)
	}

	b.ChangeAmazonURL(&u)
	u = other

	if b.AmazonURL.String() != "https://www.amazon.co.jp/dp/first" {
		t.Fatalf("AmazonURL = %s, want it unchanged by reassigning the caller's variable", b.AmazonURL)
	}
}

func TestBook_ReplaceImage(t *testing.T) {
	t.Parallel()
	b, err := book.New(mustTitle(t, "人間失格"), mustPrice(t, 1500), []author.ID{1}, book.Unpublished)
	if err != nil {
		t.Fatal(err)
	}
	first, err := book.NewImageKey("books/1/first.png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := book.NewImageKey("books/1/second.png")
	if err != nil {
		t.Fatal(err)
	}

	if prev := b.ReplaceImage(first); prev != nil {
		t.Fatalf("previous = %v, want nil for a book without an image", prev)
	}
	prev := b.ReplaceImage(second)
	if prev == nil || *prev != first {
		t.Fatalf("previous = %v, want %v", prev, first)
	}
	if b.ImageKey == nil || *b.ImageKey != second {
		t.Fatalf("ImageKey = %v, want %v", b.ImageKey, second)
	}
}

func TestBook_ChangeCatalogInfo(t *testing.T) {
	t.Parallel()
	b, err := book.New(mustTitle(t, "データ指向アプリケーションデザイン"), mustPrice(t, 4600), []author.ID{1}, book.Published)
	if err != nil {
		t.Fatal(err)
	}
	isbn, err := book.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	cover, err := book.NewCover("https://cover.openbd.jp/9784873118703.jpg", book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}

	b.ChangeCatalogInfo(&isbn, &cover)
	other, _ := book.NewISBN("9784297146221")
	isbn = other // must not leak into the book

	if b.ISBN == nil || b.ISBN.String() != "9784873118703" || b.Cover == nil || *b.Cover != cover {
		t.Fatalf("ISBN=%v Cover=%v, want the values passed in (copied)", b.ISBN, b.Cover)
	}

	b.ChangeCatalogInfo(nil, nil)
	if b.ISBN != nil || b.Cover != nil {
		t.Fatalf("ISBN=%v Cover=%v, want both cleared", b.ISBN, b.Cover)
	}
}
