package main

import (
	"context"
	"errors"
	"testing"
	"time"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestSampleBooks_AreValidAndDistinct(t *testing.T) {
	if len(sampleBooks) != 5 {
		t.Fatalf("len(sampleBooks) = %d, want 5", len(sampleBooks))
	}
	seen := map[string]bool{}
	for _, s := range sampleBooks {
		b, err := s.toBook()
		if err != nil {
			t.Fatalf("%s: %v", s.isbn, err)
		}
		if seen[b.ISBN.String()] {
			t.Fatalf("duplicate ISBN %s", b.ISBN.String())
		}
		seen[b.ISBN.String()] = true
		if b.Cover != nil {
			t.Fatalf("%s: sample data must not carry a cover (楽天由来の情報を持たない)", s.isbn)
		}
	}
}

// fakeSeedBooks は book.Repository の手書き Fake。CreateAll は createAllErr を返すか、渡された全冊を保存する。
type fakeSeedBooks struct {
	created      []*domainbook.Book
	createAllErr error
	singleCalls  int
}

func (f *fakeSeedBooks) FindByID(context.Context, domainbook.ID) (*domainbook.Book, error) {
	return nil, common.ErrNotFound
}

func (f *fakeSeedBooks) Create(context.Context, *domainbook.Book) error {
	f.singleCalls++
	return nil
}

func (f *fakeSeedBooks) CreateAll(_ context.Context, books []*domainbook.Book) error {
	if f.createAllErr != nil {
		return f.createAllErr
	}
	f.created = append(f.created, books...)
	return nil
}

func (f *fakeSeedBooks) Update(context.Context, *domainbook.Book) error { return nil }
func (f *fakeSeedBooks) Delete(context.Context, domainbook.ID) error    { return nil }

func (f *fakeSeedBooks) FindRakutenRefreshTargets(context.Context, time.Time) ([]*domainbook.Book, error) {
	return nil, nil
}

// fakeSeedQuery は book.Query の手書き Fake。total 冊の本がある DB を表す。
type fakeSeedQuery struct {
	total int
	err   error
}

func (f *fakeSeedQuery) FindDetailByID(context.Context, domainbook.ID) (*domainbook.BookDetail, error) {
	return nil, common.ErrNotFound
}

func (f *fakeSeedQuery) FindList(context.Context, domainbook.ListCondition, common.ListRange) (*domainbook.BookList, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &domainbook.BookList{Items: []*domainbook.BookListItem{}, Total: f.total}, nil
}

// fakeSeedCatalog は book.BookCatalog の手書き Fake。err が nil なら書影付きで返す。
type fakeSeedCatalog struct {
	err    error
	called int
}

func (f *fakeSeedCatalog) Lookup(_ context.Context, isbn domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	f.called++
	if f.err != nil {
		return nil, f.err
	}
	cover, err := domainbook.NewCover("https://cover.openbd.jp/"+isbn.String()+".jpg", domainbook.CoverSourceOpenBD)
	if err != nil {
		return nil, err
	}
	return &domainbook.CatalogEntry{ISBN: isbn, Cover: &cover}, nil
}

func TestSeedIfEmpty(t *testing.T) {
	ctx := context.Background()

	t.Run("本が無ければ全冊を書影付きでまとめて保存する", func(t *testing.T) {
		books := &fakeSeedBooks{}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{}, &fakeSeedCatalog{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.created) != len(sampleBooks) || books.singleCalls != 0 {
			t.Fatalf("created %d books (single Create calls: %d), want %d via CreateAll only", len(books.created), books.singleCalls, len(sampleBooks))
		}
		for _, b := range books.created {
			if b.Cover == nil || b.Cover.URL() != "https://cover.openbd.jp/"+b.ISBN.String()+".jpg" {
				t.Fatalf("%s: cover = %+v, want the catalog's cover", b.ISBN.String(), b.Cover)
			}
		}
	})

	t.Run("本が1冊でもあれば保存せず、カタログも呼ばない", func(t *testing.T) {
		books := &fakeSeedBooks{}
		catalog := &fakeSeedCatalog{}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{total: 1}, catalog); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.created) != 0 || catalog.called != 0 {
			t.Fatalf("created=%d catalog called=%d, want neither", len(books.created), catalog.called)
		}
	})

	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "カタログの障害", err: errors.New("openbd: timeout")},
		{name: "カタログに該当なし", err: common.ErrNotFound},
	} {
		t.Run(tt.name+"でも書影なしで全冊を保存する", func(t *testing.T) {
			books := &fakeSeedBooks{}
			if err := seedIfEmpty(ctx, books, &fakeSeedQuery{}, &fakeSeedCatalog{err: tt.err}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(books.created) != len(sampleBooks) {
				t.Fatalf("created %d books, want %d", len(books.created), len(sampleBooks))
			}
			for _, b := range books.created {
				if b.Cover != nil {
					t.Fatalf("%s: cover = %+v, want nil", b.ISBN.String(), b.Cover)
				}
			}
		})
	}

	t.Run("同時に起動した別のプロセスが先に入れていたら（Conflict）投入済みとして正常に終わる", func(t *testing.T) {
		books := &fakeSeedBooks{createAllErr: common.ErrConflict}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{}, &fakeSeedCatalog{}); err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
	})

	t.Run("保存のその他のエラーは返す", func(t *testing.T) {
		wantErr := errors.New("db: connection reset")
		books := &fakeSeedBooks{createAllErr: wantErr}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{}, &fakeSeedCatalog{}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("1冊でも検証を通らなければ何も保存せずエラーを返す", func(t *testing.T) {
		orig := sampleBooks
		t.Cleanup(func() { sampleBooks = orig })
		sampleBooks = append(append([]sampleBook{}, orig...), sampleBook{isbn: "123", title: "seed-test-不正", summary: "x", comment: "x", rating: 3})
		books := &fakeSeedBooks{}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{}, &fakeSeedCatalog{}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if len(books.created) != 0 {
			t.Fatalf("created %d books, want 0 (must not leave a half-seeded DB)", len(books.created))
		}
	})

	t.Run("本の有無を確かめられなければエラーを返し何も保存しない", func(t *testing.T) {
		wantErr := errors.New("db: down")
		books := &fakeSeedBooks{}
		if err := seedIfEmpty(ctx, books, &fakeSeedQuery{err: wantErr}, &fakeSeedCatalog{}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if len(books.created) != 0 {
			t.Fatalf("created %d books, want 0", len(books.created))
		}
	})
}
