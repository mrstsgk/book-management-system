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

func TestQuery_FindDetailByID_IncludesAmazonURLAndImageKey(t *testing.T) {
	db := connectTestDB(t)
	a := seedAuthor(t, db, "query-test-media", nil)
	u, err := domainbook.NewAmazonURL("https://www.amazon.co.jp/dp/4101006059")
	if err != nil {
		t.Fatal(err)
	}
	key, err := domainbook.NewImageKey("books/test/cover.png")
	if err != nil {
		t.Fatal(err)
	}
	b := newBook(t, "query-test-media", 100, []domainauthor.ID{a}, domainbook.Unpublished)
	b.ChangeAmazonURL(&u)
	b.ReplaceImage(key)
	createBook(t, db, b)

	got, err := pgbook.NewQuery(db).FindDetailByID(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("FindDetailByID: %v", err)
	}
	// The read model carries plain strings, and ImageURL stays empty: URLs are issued by the use case.
	if got.AmazonURL == nil || *got.AmazonURL != u.String() || got.ImageKey == nil || *got.ImageKey != key.String() || got.ImageURL != nil {
		t.Fatalf("url=%v key=%v imageURL=%v, want %s / %s / nil", got.AmazonURL, got.ImageKey, got.ImageURL, u, key)
	}
}
