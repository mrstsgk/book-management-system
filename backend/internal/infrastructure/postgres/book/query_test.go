package book_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
)

func TestQuery_FindDetailByID(t *testing.T) {
	db := connectTestDB(t)
	birth := time.Date(1909, 6, 19, 0, 0, 0, 0, time.UTC)
	a1 := seedAuthor(t, db, "query-test-太宰治", &birth)
	a2 := seedAuthor(t, db, "query-test-芥川龍之介", nil)
	b := newBook(t, "query-test-人間失格", 1500, []domainauthor.ID{a2, a1}, domainbook.Published)
	createBook(t, db, b)
	qry := pgbook.NewQuery(db)

	got, err := qry.FindDetailByID(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("FindDetailByID: %v", err)
	}
	birthStr := "1909-06-19"
	want := &domainbook.BookDetail{
		ID: b.ID, Title: "query-test-人間失格", Price: 1500, Status: 2, Version: 1,
		Authors: []domainbook.AuthorSummary{
			{ID: a1, Name: "query-test-太宰治", BirthDate: &birthStr, Version: 1},
			{ID: a2, Name: "query-test-芥川龍之介", BirthDate: nil, Version: 1},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}

	if _, err := qry.FindDetailByID(context.Background(), domainbook.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("unknown id: err = %v, want ErrNotFound", err)
	}
}

func TestQuery_FindSummariesByAuthorID(t *testing.T) {
	db := connectTestDB(t)
	a1 := seedAuthor(t, db, "query-test-s1", nil)
	a2 := seedAuthor(t, db, "query-test-s2", nil)
	lonely := seedAuthor(t, db, "query-test-lonely", nil)
	first := newBook(t, "query-test-first", 100, []domainauthor.ID{a1, a2}, domainbook.Unpublished)
	createBook(t, db, first)
	second := newBook(t, "query-test-second", 0, []domainauthor.ID{a1}, domainbook.Published)
	createBook(t, db, second)
	other := newBook(t, "query-test-other", 300, []domainauthor.ID{a2}, domainbook.Unpublished)
	createBook(t, db, other)
	qry := pgbook.NewQuery(db)

	t.Run("著者に紐づく書籍だけをID順に返す", func(t *testing.T) {
		got, err := qry.FindSummariesByAuthorID(context.Background(), a1)
		if err != nil {
			t.Fatalf("FindSummariesByAuthorID: %v", err)
		}
		want := []*domainbook.BookSummary{
			{ID: first.ID, Title: "query-test-first", Price: 100, Status: 1},
			{ID: second.ID, Title: "query-test-second", Price: 0, Status: 2},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("書籍が無い著者は空配列", func(t *testing.T) {
		got, err := qry.FindSummariesByAuthorID(context.Background(), lonely)
		if err != nil {
			t.Fatalf("FindSummariesByAuthorID: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("got %#v, want an empty non-nil slice", got)
		}
	})
}

func TestQuery_FindDetailByID_IncludesCatalogInfo(t *testing.T) {
	db := connectTestDB(t)
	a := seedAuthor(t, db, "query-test-catalog", nil)
	qry := pgbook.NewQuery(db)

	tests := []struct {
		name          string
		isbn          string
		wantAmazonURL *string
	}{
		{name: "978のISBNからAmazonのリンクを導出する", isbn: "9780000000002", wantAmazonURL: strPtr("https://www.amazon.co.jp/dp/0000000000")},
		{name: "979のISBNにはAmazonのリンクが無い", isbn: "9791032305690", wantAmazonURL: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isbn, err := domainbook.NewISBN(tt.isbn)
			if err != nil {
				t.Fatal(err)
			}
			cover, err := domainbook.NewCover("https://cover.openbd.jp/"+tt.isbn+".jpg", domainbook.CoverSourceOpenBD)
			if err != nil {
				t.Fatal(err)
			}
			b := newBook(t, "query-test-catalog-"+tt.isbn, 100, []domainauthor.ID{a}, domainbook.Unpublished)
			b.ChangeCatalogInfo(&isbn, &cover)
			createBook(t, db, b)

			got, err := qry.FindDetailByID(context.Background(), b.ID)
			if err != nil {
				t.Fatalf("FindDetailByID: %v", err)
			}
			// The read model carries plain strings, not the write-side VOs.
			if got.ISBN == nil || *got.ISBN != tt.isbn || got.CoverURL == nil || *got.CoverURL != cover.URL() ||
				got.CoverSource == nil || *got.CoverSource != "openbd" {
				t.Fatalf("isbn=%v cover=%v source=%v", got.ISBN, got.CoverURL, got.CoverSource)
			}
			if (got.AmazonURL == nil) != (tt.wantAmazonURL == nil) || (got.AmazonURL != nil && *got.AmazonURL != *tt.wantAmazonURL) {
				t.Fatalf("AmazonURL = %v, want %v", got.AmazonURL, tt.wantAmazonURL)
			}
		})
	}

	t.Run("ISBNが無ければAmazonのリンクも書影も無い", func(t *testing.T) {
		b := newBook(t, "query-test-catalog-none", 100, []domainauthor.ID{a}, domainbook.Unpublished)
		createBook(t, db, b)
		got, err := qry.FindDetailByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.ISBN != nil || got.AmazonURL != nil || got.CoverURL != nil || got.CoverSource != nil {
			t.Fatalf("got %+v, want no catalog info", got)
		}
	})
}

func strPtr(s string) *string { return &s }
