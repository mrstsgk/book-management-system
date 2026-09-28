package tag_test

import (
	"context"
	"reflect"
	"testing"

	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
)

func TestQuery_FindList(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	q := pgtag.NewQuery(db)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'query-test-%'") })

	a := createTag(t, repo, "query-test-データベース")
	b := createTag(t, repo, "query-test-分散システム")

	list, err := q.FindList(context.Background())
	if err != nil {
		t.Fatalf("FindList: %v", err)
	}
	items := map[domaintag.ID]*domaintag.TagListItem{}
	for _, it := range list.Items {
		items[it.ID] = it
	}
	if items[a.ID] == nil || items[a.ID].Name != "query-test-データベース" || items[a.ID].Version != 1 {
		t.Fatalf("got %+v, want Name=query-test-データベース Version=1", items[a.ID])
	}
	if items[b.ID] == nil || items[b.ID].Name != "query-test-分散システム" || items[b.ID].Version != 1 {
		t.Fatalf("got %+v, want Name=query-test-分散システム Version=1", items[b.ID])
	}
}

func TestQuery_ExistsAll(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	q := pgtag.NewQuery(db)
	t.Cleanup(func() { db.Exec("DELETE FROM tag WHERE name LIKE 'query-test-%'") })

	a := createTag(t, repo, "query-test-存在確認A")
	b := createTag(t, repo, "query-test-存在確認B")

	tests := []struct {
		name string
		ids  []domaintag.ID
		want bool
	}{
		{name: "空配列は常にtrue", ids: []domaintag.ID{}, want: true},
		{name: "全部実在すればtrue", ids: []domaintag.ID{a.ID, b.ID}, want: true},
		{name: "1つでも欠ければfalse", ids: []domaintag.ID{a.ID, domaintag.ID(-1)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := q.ExistsAll(context.Background(), tt.ids)
			if err != nil {
				t.Fatalf("ExistsAll: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_CountBooks(t *testing.T) {
	db := connectTestDB(t)
	repo := pgtag.NewRepository(db)
	q := pgtag.NewQuery(db)
	isbns := []string{"9780000003201", "9780000003218", "9780000003225"}
	cleanup := func() {
		db.Exec("DELETE FROM book WHERE isbn IN ?", isbns)
		db.Exec("DELETE FROM tag WHERE name LIKE 'counts-test-%'")
	}
	cleanup()
	t.Cleanup(cleanup)

	a := createTag(t, repo, "counts-test-a")
	b1 := createTag(t, repo, "counts-test-b1")
	b2 := createTag(t, repo, "counts-test-b2")
	c := createTag(t, repo, "counts-test-c")
	d := createTag(t, repo, "counts-test-d")
	insertBook := func(isbn string, tags ...*domaintag.Tag) {
		t.Helper()
		var id int64
		if err := db.Raw("INSERT INTO book (isbn, title, summary, comment, rating, version) VALUES (?, 'counts-test-書名', 'まとめ', '感想', 3, 1) RETURNING id", isbn).Scan(&id).Error; err != nil {
			t.Fatalf("insert book: %v", err)
		}
		for _, tg := range tags {
			if err := db.Exec("INSERT INTO book_tag (book_id, tag_id) VALUES (?, ?)", id, int64(tg.ID)).Error; err != nil {
				t.Fatalf("insert book_tag: %v", err)
			}
		}
	}
	// 同じ本に複数のタグが付いていても、各タグではその本を1冊として数える
	insertBook(isbns[0], a, b1)
	insertBook(isbns[1], a, b2)
	insertBook(isbns[2], d)
	if err := repo.Delete(context.Background(), d.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := q.CountBooks(context.Background())
	if err != nil {
		t.Fatalf("CountBooks: %v", err)
	}
	// 共有DBに他のテストのタグが残っていても通るよう、自分が作ったタグの行だけを取り出す
	mine := map[domaintag.ID]bool{a.ID: true, b1.ID: true, b2.ID: true, c.ID: true, d.ID: true}
	var ids []domaintag.ID
	var counts []int
	var names []string
	for _, it := range got.Items {
		if mine[it.ID] {
			ids = append(ids, it.ID)
			counts = append(counts, it.BookCount)
			names = append(names, it.Name)
		}
	}
	// 冊数の多い順、同数ならタグ名順。0冊のタグ（c）と削除したタグ（d）は含めない
	if want := []domaintag.ID{a.ID, b1.ID, b2.ID}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	if !reflect.DeepEqual(counts, []int{2, 1, 1}) {
		t.Fatalf("counts = %v, want [2 1 1]", counts)
	}
	if !reflect.DeepEqual(names, []string{"counts-test-a", "counts-test-b1", "counts-test-b2"}) {
		t.Fatalf("names = %v", names)
	}
}
