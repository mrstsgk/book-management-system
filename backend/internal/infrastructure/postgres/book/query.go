package book

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type query struct {
	db  *gorm.DB
	now func() time.Time
}

func NewQuery(db *gorm.DB, now func() time.Time) domainbook.Query {
	return &query{db: db, now: now}
}

// visibleCover は画面に出してよい書影（URL・提供元・楽天の商品ページ）を返す。楽天の書影が保持期限を過ぎていれば
// 3つとも nil にする（取り直し・消去が走る前でも、期限切れの楽天由来の情報を返さないため）。
func visibleCover(row model, now time.Time) (url, source, productURL *string) {
	if row.CoverSource != nil && domainbook.CoverSource(*row.CoverSource) == domainbook.CoverSourceRakuten &&
		(row.CoverFetchedAt == nil || domainbook.RakutenExpired(*row.CoverFetchedAt, now)) {
		return nil, nil, nil
	}
	return row.CoverURL, row.CoverSource, row.CoverProductURL
}

func (q *query) FindDetailByID(ctx context.Context, id domainbook.ID) (*domainbook.BookDetail, error) {
	var row model
	if err := q.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 読んだ本が見つかりません", common.ErrNotFound)
		}
		return nil, err
	}
	tagNames, err := tagNamesByBookID(ctx, q.db, []int64{row.ID})
	if err != nil {
		return nil, err
	}
	coverURL, coverSource, coverProductURL := visibleCover(row, q.now())
	return &domainbook.BookDetail{
		ID: domainbook.ID(row.ID), ISBN: row.ISBN, Title: displayTitle(row.TitleOverride, row.Title), Authors: row.Authors,
		Publisher: row.Publisher, PublishedOn: row.PublishedOn, AmazonURL: amazonURLOf(row.ISBN),
		CoverURL: coverURL, CoverSource: coverSource, CoverProductURL: coverProductURL, TitleOverride: row.TitleOverride, Summary: row.Summary,
		Tags: orEmpty(tagNames[row.ID]), Comment: row.Comment, Rating: row.Rating, Version: row.Version,
	}, nil
}

// displayTitle は画面に出す書名を決める。自分で上書きした書名があればそれ、無ければ外部カタログの書名。
func displayTitle(override *string, catalogTitle string) string {
	if override != nil {
		return *override
	}
	return catalogTitle
}

// bookTagRow は book_tag と tag を結合した1行（どの本にどのタグ名が付くか）。
type bookTagRow struct {
	BookID  int64
	TagName string
}

// tagNamesByBookID は bookIDs に対応するタグ名を book_id ごとにまとめて返す（1回の別クエリ、N+1にしない）。
func tagNamesByBookID(ctx context.Context, db *gorm.DB, bookIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(bookIDs))
	if len(bookIDs) == 0 {
		return result, nil
	}
	var rows []bookTagRow
	err := db.WithContext(ctx).Table("book_tag").
		Select("book_tag.book_id AS book_id, tag.name AS tag_name").
		Joins("JOIN tag ON tag.id = book_tag.tag_id").
		Where("book_tag.book_id IN ?", bookIDs).
		Order("tag.name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.BookID] = append(result[row.BookID], row.TagName)
	}
	return result, nil
}

// orEmpty は nil のスライスを空スライスに揃える（Read Model は「タグ無し」と「未取得」を区別しないため）。
func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
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
	err := db.Select("id, isbn, title, title_override, authors, cover_url, cover_source, cover_product_url, cover_fetched_at, summary, rating").
		Order("created_at DESC, id DESC").Limit(r.Limit()).Offset(r.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	bookIDs := make([]int64, len(rows))
	for i, row := range rows {
		bookIDs[i] = row.ID
	}
	tagNames, err := tagNamesByBookID(ctx, q.db, bookIDs)
	if err != nil {
		return nil, err
	}
	list := &domainbook.BookList{Items: make([]*domainbook.BookListItem, 0, len(rows)), Total: int(total)}
	now := q.now()
	for _, row := range rows {
		coverURL, coverSource, coverProductURL := visibleCover(row, now)
		list.Items = append(list.Items, &domainbook.BookListItem{
			ID: domainbook.ID(row.ID), ISBN: row.ISBN, Title: displayTitle(row.TitleOverride, row.Title), Summary: row.Summary,
			Authors: row.Authors, AmazonURL: amazonURLOf(row.ISBN), CoverURL: coverURL, CoverSource: coverSource, CoverProductURL: coverProductURL,
			Rating: row.Rating, Tags: orEmpty(tagNames[row.ID]),
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
