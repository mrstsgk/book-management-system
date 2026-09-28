package book

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type query struct {
	db *gorm.DB
}

func NewQuery(db *gorm.DB) domainbook.Query {
	return &query{db: db}
}

func (q *query) FindDetailByID(ctx context.Context, id domainbook.ID) (*domainbook.BookDetail, error) {
	var row model
	if err := q.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 読んだ本が見つかりません", common.ErrNotFound)
		}
		return nil, err
	}
	return &domainbook.BookDetail{
		ID: domainbook.ID(row.ID), ISBN: row.ISBN, Title: row.Title, Authors: row.Authors,
		Publisher: row.Publisher, PublishedOn: row.PublishedOn, AmazonURL: amazonURLOf(row.ISBN),
		CoverURL: row.CoverURL, CoverSource: row.CoverSource, Comment: row.Comment, Rating: row.Rating,
		Version: row.Version,
	}, nil
}

// FindList は新しく登録した順（同時刻は ID の大きい順）に、取得範囲の分だけ返す。総件数も返す。
func (q *query) FindList(ctx context.Context, r common.ListRange) (*domainbook.BookList, error) {
	db := q.db.WithContext(ctx)
	// 総件数と取得範囲の取得は別の SQL なので、同時に登録されると1件ずれうる。
	// 一覧画面では許容でき、スナップショットのトランザクションより軽い。
	var total int64
	if err := db.Model(&model{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model
	err := db.Select("id, isbn, title, authors, cover_url, cover_source, rating").
		Order("created_at DESC, id DESC").Limit(r.Limit()).Offset(r.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	list := &domainbook.BookList{Items: make([]*domainbook.BookListItem, 0, len(rows)), Total: int(total)}
	for _, row := range rows {
		list.Items = append(list.Items, &domainbook.BookListItem{
			ID: domainbook.ID(row.ID), ISBN: row.ISBN, Title: row.Title, Authors: row.Authors,
			AmazonURL: amazonURLOf(row.ISBN), CoverURL: row.CoverURL, CoverSource: row.CoverSource, Rating: row.Rating,
		})
	}
	return list, nil
}

// amazonURLOf は保存済みの ISBN から商品ページのリンクを導出する（domainbook.ISBN.AmazonURL）。
// 保存済みの ISBN は書き込み時に検証済みなので、解釈できなければリンクなしとする。
func amazonURLOf(isbn string) *string {
	v, err := domainbook.NewISBN(isbn)
	if err != nil {
		return nil
	}
	u, ok := v.AmazonURL()
	if !ok {
		return nil
	}
	return &u
}
