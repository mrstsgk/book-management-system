package tag

import (
	"context"

	"gorm.io/gorm"

	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type query struct {
	db *gorm.DB
}

func NewQuery(db *gorm.DB) domaintag.Query {
	return &query{db: db}
}

// FindList は全件を id 昇順で返す。
func (q *query) FindList(ctx context.Context) (*domaintag.TagList, error) {
	var rows []model
	if err := q.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := &domaintag.TagList{Items: make([]*domaintag.TagListItem, 0, len(rows))}
	for _, row := range rows {
		list.Items = append(list.Items, &domaintag.TagListItem{ID: domaintag.ID(row.ID), Name: row.Name, Version: row.Version})
	}
	return list, nil
}

// CountBooks は tag と book_tag を内部結合して数える（結合しない tag、つまり0冊のタグは結果に出ない）。
func (q *query) CountBooks(ctx context.Context) (*domaintag.TagBookCounts, error) {
	var rows []struct {
		ID        int64
		Name      string
		BookCount int
	}
	err := q.db.WithContext(ctx).Table("tag").
		Select("tag.id AS id, tag.name AS name, COUNT(*) AS book_count").
		Joins("JOIN book_tag ON book_tag.tag_id = tag.id").
		Group("tag.id, tag.name").
		Order("book_count DESC, tag.name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := &domaintag.TagBookCounts{Items: make([]*domaintag.TagBookCount, 0, len(rows))}
	for _, row := range rows {
		counts.Items = append(counts.Items, &domaintag.TagBookCount{ID: domaintag.ID(row.ID), Name: row.Name, BookCount: row.BookCount})
	}
	return counts, nil
}

// ExistsAll は ids がすべて実在するかを、一致件数が ids の個数と等しいかで判定する。
func (q *query) ExistsAll(ctx context.Context, ids []domaintag.ID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	raw := make([]int64, len(ids))
	for i, id := range ids {
		raw[i] = int64(id)
	}
	var count int64
	if err := q.db.WithContext(ctx).Model(&model{}).Where("id IN ?", raw).Count(&count).Error; err != nil {
		return false, err
	}
	return int(count) == len(ids), nil
}
