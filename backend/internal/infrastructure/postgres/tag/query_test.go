package tag_test

import (
	"context"
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
