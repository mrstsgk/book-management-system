package postgres

import (
	"context"

	"gorm.io/gorm"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// query は、分野タグの一覧・集計を PostgreSQL から読み取る（tagdomain.Query の実装）。
type query struct {
	db *gorm.DB
}

// NewQuery は、db を使う Query を作る。
func NewQuery(db *gorm.DB) tagdomain.Query {
	return &query{db: db}
}

// FindAll は全件を id 昇順で返す。
func (q *query) FindAll(ctx context.Context) (*tagdomain.TagList, error) {
	var rows []model
	if err := q.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	list := &tagdomain.TagList{Items: make([]*tagdomain.TagListItem, 0, len(rows))}
	for _, row := range rows {
		list.Items = append(list.Items, &tagdomain.TagListItem{ID: tagdomain.ID(row.ID), Name: row.Name, Version: row.Version})
	}
	return list, nil
}

// CountBooks は tag と book_tag を内部結合して数える（結合しない tag、つまり0冊のタグは結果に出ない）。
func (q *query) CountBooks(ctx context.Context) (*tagdomain.TagBookCounts, error) {
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
	counts := &tagdomain.TagBookCounts{Items: make([]*tagdomain.TagBookCount, 0, len(rows))}
	for _, row := range rows {
		counts.Items = append(counts.Items, &tagdomain.TagBookCount{ID: tagdomain.ID(row.ID), Name: row.Name, BookCount: row.BookCount})
	}
	return counts, nil
}

// ExistsAll は ids がすべて実在するかを、一致件数が ids の個数と等しいかで判定する。
func (q *query) ExistsAll(ctx context.Context, ids []tagdomain.ID) (bool, error) {
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

// ExistsByName は name と同じ名前の行（excludeID の行を除く）が 1 件以上あるかを返す。比較は保存された名前と完全一致。
func (q *query) ExistsByName(ctx context.Context, name string, excludeID tagdomain.ID) (bool, error) {
	var count int64
	tx := q.db.WithContext(ctx).Model(&model{}).Where("name = ?", name)
	if excludeID != 0 {
		tx = tx.Where("id <> ?", int64(excludeID))
	}
	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
