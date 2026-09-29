package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

var refreshNow = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

// fakeRefreshBooks は取り直しの対象と、本ごとの保存結果を返せる book.Repository の Fake。
type fakeRefreshBooks struct {
	fakeBooks
	targets       []*book.Book
	targetsErr    error
	gotBefore     time.Time
	updateErrByID map[book.ID]error
	updated       []*book.Book
}

func (f *fakeRefreshBooks) FindCoverRefreshTargets(_ context.Context, fetchedBefore time.Time) ([]*book.Book, error) {
	f.gotBefore = fetchedBefore
	return f.targets, f.targetsErr
}

func (f *fakeRefreshBooks) Update(_ context.Context, b *book.Book) error {
	if err := f.updateErrByID[b.ID]; err != nil {
		return err
	}
	f.updated = append(f.updated, b)
	return nil
}

// fakeCatalogByISBN は ISBN ごとに結果を返す book.BookCatalog の Fake。
type fakeCatalogByISBN struct {
	entries map[string]*book.CatalogEntry
	err     error
	// cancelOnCall が設定されていれば、その回数目の呼び出しの直前に cancel を呼ぶ
	// （ループ処理中にctxがキャンセルされた状況を模す）。
	cancelOnCall int
	cancel       func()
	calls        int
}

func (f *fakeCatalogByISBN) Lookup(_ context.Context, isbn book.ISBN) (*book.CatalogEntry, error) {
	f.calls++
	if f.cancelOnCall != 0 && f.calls == f.cancelOnCall {
		f.cancel()
	}
	if f.err != nil {
		return nil, f.err
	}
	if e, ok := f.entries[isbn.String()]; ok {
		return e, nil
	}
	return nil, common.ErrNotFound
}

func rakutenBook(t *testing.T, id book.ID, isbn string, fetchedAt time.Time) *book.Book {
	t.Helper()
	i, err := book.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, _ := book.NewBibliography("書名", "著者", "出版社", "202601")
	cover, err := book.NewRakutenCover("https://thumbnail.image.rakuten.co.jp/old.jpg", "https://books.rakuten.co.jp/rb/old/", fetchedAt)
	if err != nil {
		t.Fatal(err)
	}
	summary, _ := book.NewSummary("まとめ")
	comment, _ := book.NewComment("感想")
	rating, _ := book.NewRating(3)
	b := book.New(i, bib, &cover, summary, comment, rating, book.TagSelection{})
	b.ID, b.Version = id, 1
	return b
}

func newRakutenCover(t *testing.T, fetchedAt time.Time) *book.Cover {
	t.Helper()
	c, err := book.NewRakutenCover("https://thumbnail.image.rakuten.co.jp/new.jpg", "https://books.rakuten.co.jp/rb/new/", fetchedAt)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func TestRefreshRakutenCoversUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("取り直す時期の基準として今から83日前を渡す", func(t *testing.T) {
		t.Parallel()
		books := &fakeRefreshBooks{}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: &fakeCatalogByISBN{}, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := refreshNow.Add(-book.RakutenRefreshAfter); !books.gotBefore.Equal(want) {
			t.Fatalf("fetchedBefore = %v, want %v", books.gotBefore, want)
		}
	})

	t.Run("取り直せたら新しい書影で保存する", func(t *testing.T) {
		t.Parallel()
		b := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRefreshAfter))
		fresh := newRakutenCover(t, refreshNow)
		entry := catalogEntry(t, "取り直した書名", fresh)
		books := &fakeRefreshBooks{targets: []*book.Book{b}}
		catalog := &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{"9780000003652": entry}}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updated) != 1 || books.updated[0].Cover == nil || *books.updated[0].Cover != *fresh {
			t.Fatalf("updated %+v, want the book saved with the refreshed cover", books.updated)
		}
	})

	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "障害", err: errors.New("rakuten: timeout")},
		{name: "該当なし", err: common.ErrNotFound},
	} {
		t.Run("取り直せず（"+tt.name+"）期限切れなら書影を外して保存する", func(t *testing.T) {
			t.Parallel()
			b := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention))
			books := &fakeRefreshBooks{targets: []*book.Book{b}}
			uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: &fakeCatalogByISBN{err: tt.err}, Now: func() time.Time { return refreshNow }}

			if err := uc.Execute(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(books.updated) != 1 || books.updated[0].Cover != nil {
				t.Fatalf("updated %+v, want the book saved without the expired cover", books.updated)
			}
		})
	}

	t.Run("取り直せず期限前なら保存しない", func(t *testing.T) {
		t.Parallel()
		b := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention+time.Second))
		books := &fakeRefreshBooks{targets: []*book.Book{b}}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: &fakeCatalogByISBN{err: errors.New("rakuten: timeout")}, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updated) != 0 || b.Cover == nil {
			t.Fatalf("updated %+v cover=%v, want nothing saved and the cover kept until it expires", books.updated, b.Cover)
		}
	})

	t.Run("再取得できたが書影が無く期限前なら、今の楽天の書影を残し保存しない", func(t *testing.T) {
		t.Parallel()
		b := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention+time.Second))
		entry := catalogEntry(t, "取り直した書名", nil)
		books := &fakeRefreshBooks{targets: []*book.Book{b}}
		catalog := &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{"9780000003652": entry}}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updated) != 0 || b.Cover == nil {
			t.Fatalf("updated %+v cover=%v, want the still-valid rakuten cover kept and nothing saved", books.updated, b.Cover)
		}
	})

	t.Run("再取得できたが書影が無く期限切れなら、今の楽天の書影を外して保存する", func(t *testing.T) {
		t.Parallel()
		b := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention))
		entry := catalogEntry(t, "取り直した書名", nil)
		books := &fakeRefreshBooks{targets: []*book.Book{b}}
		catalog := &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{"9780000003652": entry}}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updated) != 1 || books.updated[0].Cover != nil {
			t.Fatalf("updated %+v, want the expired rakuten cover dropped and saved", books.updated)
		}
	})

	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "競合（その間に自分が更新した）", err: common.ErrConflict},
		{name: "その他のエラー", err: errors.New("db: connection reset")},
	} {
		t.Run("1冊目の保存が"+tt.name+"でも2冊目を処理する", func(t *testing.T) {
			t.Parallel()
			first := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention))
			second := rakutenBook(t, 2, "9780000003669", refreshNow.Add(-book.RakutenRetention))
			books := &fakeRefreshBooks{targets: []*book.Book{first, second}, updateErrByID: map[book.ID]error{1: tt.err}}
			uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: &fakeCatalogByISBN{err: common.ErrNotFound}, Now: func() time.Time { return refreshNow }}

			if err := uc.Execute(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(books.updated) != 1 || books.updated[0].ID != 2 {
				t.Fatalf("updated %+v, want only the second book saved", books.updated)
			}
		})
	}

	t.Run("対象の読み込みに失敗したらエラーを返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: connection refused")
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: &fakeRefreshBooks{targetsErr: wantErr}, Catalog: &fakeCatalogByISBN{}, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("処理の途中でctxがキャンセルされたら残りを処理せず止まる", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		first := rakutenBook(t, 1, "9780000003652", refreshNow.Add(-book.RakutenRetention))
		second := rakutenBook(t, 2, "9780000003669", refreshNow.Add(-book.RakutenRetention))
		third := rakutenBook(t, 3, "9780000003676", refreshNow.Add(-book.RakutenRetention))
		books := &fakeRefreshBooks{targets: []*book.Book{first, second, third}}
		// 1冊目の処理中にctxがキャンセルされた状況を模す
		catalog := &fakeCatalogByISBN{err: errors.New("rakuten: timeout"), cancelOnCall: 1, cancel: cancel}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if catalog.calls != 1 {
			t.Fatalf("Catalog.Lookup was called %d times, want 1 (stop once ctx is cancelled)", catalog.calls)
		}
	})
}

func TestRefreshRakutenCoversUsecase_Execute_CoverlessBook(t *testing.T) {
	t.Parallel()
	const isbn = "9780000005007"
	coverless := func(t *testing.T) *book.Book {
		b := rakutenBook(t, 1, isbn, refreshNow)
		b.Cover = nil
		return b
	}
	openBDCover, err := book.NewCover("https://cover.openbd.jp/"+isbn+".jpg", book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name  string
		cover *book.Cover
	}{
		{name: "openBDの書影", cover: &openBDCover},
		{name: "楽天の書影", cover: newRakutenCover(t, refreshNow)},
	} {
		t.Run("書影なしの本に"+tt.name+"が取れたら、その書影で保存する", func(t *testing.T) {
			t.Parallel()
			books := &fakeRefreshBooks{targets: []*book.Book{coverless(t)}}
			catalog := &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{isbn: catalogEntry(t, "書名", tt.cover)}}
			uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

			if err := uc.Execute(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(books.updated) != 1 || books.updated[0].Cover == nil || *books.updated[0].Cover != *tt.cover {
				t.Fatalf("updated %+v, want the book saved with the found cover", books.updated)
			}
		})
	}

	for _, tt := range []struct {
		name    string
		catalog *fakeCatalogByISBN
	}{
		{name: "該当はあるが書影が無い", catalog: &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{isbn: catalogEntry(t, "書名", nil)}}},
		{name: "カタログに該当なし", catalog: &fakeCatalogByISBN{}},
		{name: "カタログの障害", catalog: &fakeCatalogByISBN{err: errors.New("openbd: timeout")}},
	} {
		t.Run("書影なしの本で"+tt.name+"なら保存しない", func(t *testing.T) {
			t.Parallel()
			b := coverless(t)
			books := &fakeRefreshBooks{targets: []*book.Book{b}}
			uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: tt.catalog, Now: func() time.Time { return refreshNow }}

			if err := uc.Execute(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(books.updated) != 0 || b.Cover != nil || b.Version != 1 {
				t.Fatalf("updated=%d cover=%+v version=%d, want the book left unchanged", len(books.updated), b.Cover, b.Version)
			}
		})
	}

	t.Run("楽天の削除指示を受けた本に楽天の書影が来ても付けない", func(t *testing.T) {
		t.Parallel()
		b := coverless(t)
		b.DisableRakuten()
		books := &fakeRefreshBooks{targets: []*book.Book{b}}
		catalog := &fakeCatalogByISBN{entries: map[string]*book.CatalogEntry{isbn: catalogEntry(t, "書名", newRakutenCover(t, refreshNow))}}
		uc := &command.RefreshRakutenCoversUsecaseImpl{Books: books, Catalog: catalog, Now: func() time.Time { return refreshNow }}

		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, u := range books.updated {
			if u.Cover != nil {
				t.Fatalf("saved cover %+v, want no Rakuten cover on a Rakuten-disabled book", u.Cover)
			}
		}
	})
}
