package book_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
)

func strPtr(s string) *string { return &s }

func TestQuery_FindDetailByID(t *testing.T) {
	db := connectTestDB(t)
	qry := pgbook.NewQuery(db)

	tests := []struct {
		name          string
		isbn          string
		cover         *domainbook.Cover
		wantAmazonURL *string
	}{
		{name: "978のISBNからAmazonのリンクを導出し書影も返す", isbn: "9780000000057", cover: mustCover(t, "https://cover.openbd.jp/q.jpg"), wantAmazonURL: strPtr("https://www.amazon.co.jp/dp/0000000051")},
		{name: "979のISBNはAmazonのリンクなし、書影なしはnull", isbn: "9791032305690", cover: nil, wantAmazonURL: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBook(t, tt.isbn, "query-test-書名", tt.cover, 4)
			createBook(t, db, b)

			got, err := qry.FindDetailByID(context.Background(), b.ID)
			if err != nil {
				t.Fatalf("FindDetailByID: %v", err)
			}
			want := &domainbook.BookDetail{
				ID: b.ID, ISBN: tt.isbn, Title: "query-test-書名", Authors: "Kleppmann,Martin", Publisher: "オーム社",
				PublishedOn: "201907", AmazonURL: tt.wantAmazonURL, Summary: "一言まとめ", Tags: []string{}, Comment: "感想\n2行目", Rating: 4, Version: 1,
			}
			if tt.cover != nil {
				want.CoverURL, want.CoverSource = strPtr(tt.cover.URL()), strPtr("openbd")
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v\nwant %+v", got, want)
			}
		})
	}

	if _, err := qry.FindDetailByID(context.Background(), domainbook.ID(-1)); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("unknown id: err = %v, want ErrNotFound", err)
	}
}

func TestQuery_FindList(t *testing.T) {
	db := connectTestDB(t)
	qry := pgbook.NewQuery(db)
	ctx := context.Background()
	listRange := func(t *testing.T, limit, offset int) domaincommon.ListRange {
		t.Helper()
		r, err := domaincommon.NewListRange(limit, offset)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	// 開発用 DB に既存の行があっても、ここで登録した本は新しい順の先頭に並ぶ
	older := newBook(t, "9780000000064", "list-test-older", nil, 3)
	createBook(t, db, older)
	newer := newBook(t, "9780000000071", "list-test-newer", mustCover(t, "https://cover.openbd.jp/newer.jpg"), 5)
	createBook(t, db, newer)
	all, err := qry.FindList(ctx, listRange(t, 1, 0))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("新しく登録した順に取得範囲の分を返し、感想の本文は含めない", func(t *testing.T) {
		got, err := qry.FindList(ctx, listRange(t, 2, 0))
		if err != nil {
			t.Fatalf("FindList: %v", err)
		}
		want := &domainbook.BookList{Total: all.Total, Items: []*domainbook.BookListItem{
			{ID: newer.ID, ISBN: "9780000000071", Title: "list-test-newer", Summary: "一言まとめ", Tags: []string{}, Authors: "Kleppmann,Martin",
				AmazonURL: strPtr("https://www.amazon.co.jp/dp/0000000078"), CoverURL: strPtr("https://cover.openbd.jp/newer.jpg"), CoverSource: strPtr("openbd"), Rating: 5},
			{ID: older.ID, ISBN: "9780000000064", Title: "list-test-older", Summary: "一言まとめ", Tags: []string{}, Authors: "Kleppmann,Martin",
				AmazonURL: strPtr("https://www.amazon.co.jp/dp/000000006X"), Rating: 3},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v\nwant %+v", got, want)
		}
	})

	t.Run("続きの範囲と範囲外", func(t *testing.T) {
		got, err := qry.FindList(ctx, listRange(t, 1, 1))
		if err != nil || len(got.Items) != 1 || got.Items[0].ID != older.ID {
			t.Fatalf("got (%+v, %v), want only the older book", got, err)
		}
		out, err := qry.FindList(ctx, listRange(t, 1, all.Total))
		if err != nil || out.Items == nil || len(out.Items) != 0 || out.Total != all.Total {
			t.Fatalf("got (%#v, %v), want an empty non-nil list with total %d", out, err, all.Total)
		}
	})
}

func TestQuery_DisplayTitle(t *testing.T) {
	db := connectTestDB(t)
	q := pgbook.NewQuery(db)

	overridden := newBook(t, "9780000001016", "カタログの書名", nil, 4)
	title, _ := domainbook.NewTitle("正しい書名")
	overridden.OverrideTitle(&title)
	createBook(t, db, overridden)
	plain := newBook(t, "9780000001023", "上書きなしの書名", nil, 4)
	createBook(t, db, plain)

	t.Run("詳細は上書きを優先し、上書きも返す", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), overridden.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.Title != "正しい書名" || got.TitleOverride == nil || *got.TitleOverride != "正しい書名" || got.Summary != "一言まとめ" {
			t.Fatalf("got title=%q override=%v summary=%q", got.Title, got.TitleOverride, got.Summary)
		}
	})

	t.Run("上書きが無ければカタログの書名を表示し、上書きはnil", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), plain.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.Title != "上書きなしの書名" || got.TitleOverride != nil {
			t.Fatalf("got title=%q override=%v", got.Title, got.TitleOverride)
		}
	})

	t.Run("一覧も表示する書名と一言まとめを返す", func(t *testing.T) {
		r, _ := domaincommon.NewListRange(100, 0)
		list, err := q.FindList(context.Background(), r)
		if err != nil {
			t.Fatalf("FindList: %v", err)
		}
		titles := map[domainbook.ID]*domainbook.BookListItem{}
		for _, it := range list.Items {
			titles[it.ID] = it
		}
		if it := titles[overridden.ID]; it == nil || it.Title != "正しい書名" || it.Summary != "一言まとめ" {
			t.Fatalf("overridden item = %+v", it)
		}
		if it := titles[plain.ID]; it == nil || it.Title != "上書きなしの書名" {
			t.Fatalf("plain item = %+v", it)
		}
	})
}

func TestQuery_Tags(t *testing.T) {
	db := connectTestDB(t)
	q := pgbook.NewQuery(db)

	tagA := mustCreateTag(t, db, "query-test-タグA")
	tagB := mustCreateTag(t, db, "query-test-タグB")
	tagC := mustCreateTag(t, db, "query-test-タグC")
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'query-test-%'") })

	tagged := newBook(t, "9780000002440", "タグ付きの本", nil, 4, domaintag.ID(tagA), domaintag.ID(tagB))
	createBook(t, db, tagged)
	untagged := newBook(t, "9780000002457", "タグ無しの本", nil, 4)
	createBook(t, db, untagged)
	// 別のタグが付いた本を後から登録する（一覧のページングをまたいでも他の本のタグと混ざらないことを確かめるため）。
	taggedC := newBook(t, "9780000002471", "別のタグの本", nil, 4, domaintag.ID(tagC))
	createBook(t, db, taggedC)

	t.Run("詳細はタグ名を名前順で返す", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), tagged.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		want := []string{"query-test-タグA", "query-test-タグB"}
		if !reflect.DeepEqual(got.Tags, want) {
			t.Fatalf("Tags = %v, want %v", got.Tags, want)
		}
	})

	t.Run("タグが無い本は空スライスを返す", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), untagged.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.Tags == nil || len(got.Tags) != 0 {
			t.Fatalf("Tags = %v, want an empty non-nil slice", got.Tags)
		}
	})

	t.Run("一覧はそれぞれの本に対応するタグ名だけを返す", func(t *testing.T) {
		r, err := domaincommon.NewListRange(100, 0)
		if err != nil {
			t.Fatal(err)
		}
		list, err := q.FindList(context.Background(), r)
		if err != nil {
			t.Fatalf("FindList: %v", err)
		}
		byID := map[domainbook.ID]*domainbook.BookListItem{}
		for _, it := range list.Items {
			byID[it.ID] = it
		}
		if it := byID[tagged.ID]; it == nil || !reflect.DeepEqual(it.Tags, []string{"query-test-タグA", "query-test-タグB"}) {
			t.Fatalf("tagged item = %+v", it)
		}
		if it := byID[untagged.ID]; it == nil || len(it.Tags) != 0 {
			t.Fatalf("untagged item = %+v", it)
		}
		if it := byID[taggedC.ID]; it == nil || !reflect.DeepEqual(it.Tags, []string{"query-test-タグC"}) {
			t.Fatalf("taggedC item = %+v, want only query-test-タグC (must not pick up tagged's tags)", it)
		}
	})

	t.Run("ページの境界をまたいでも他の本のタグは混ざらない", func(t *testing.T) {
		// 新しく登録した順に並ぶため、1件目は taggedC のはず。
		r, err := domaincommon.NewListRange(1, 0)
		if err != nil {
			t.Fatal(err)
		}
		list, err := q.FindList(context.Background(), r)
		if err != nil {
			t.Fatalf("FindList: %v", err)
		}
		if len(list.Items) != 1 || list.Items[0].ID != taggedC.ID {
			t.Fatalf("Items = %+v, want only taggedC (%d)", list.Items, taggedC.ID)
		}
		if got := list.Items[0].Tags; !reflect.DeepEqual(got, []string{"query-test-タグC"}) {
			t.Fatalf("Tags = %v, want only query-test-タグC (tagged's タグA/タグB must not leak in)", got)
		}
	})
}
