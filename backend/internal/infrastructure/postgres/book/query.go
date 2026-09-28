package book

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type query struct {
	db *gorm.DB
}

func NewQuery(db *gorm.DB) domainbook.Query {
	return &query{db: db}
}

func NewListQuery(db *gorm.DB) domainbook.ListQuery {
	return &query{db: db}
}

type detailRow struct {
	ID              int64
	Title           string
	Price           yen
	PublishStatus   int
	ISBN            *string
	CoverURL        *string
	CoverSource     *string
	Version         int
	AuthorID        int64
	AuthorName      string
	AuthorBirthDate *time.Time
	AuthorVersion   int
}

// FindDetailByID は書籍と紐づく著者を1クエリで取得し、書籍1件＋著者一覧の Read Model に組み立てる。
func (q *query) FindDetailByID(ctx context.Context, id domainbook.ID) (*domainbook.BookDetail, error) {
	var rows []detailRow
	// Inner joins: every book has at least one author (enforced on every write).
	err := q.db.WithContext(ctx).
		Table("book AS b").
		Select(`b.id, b.title, b.price, b.publish_status, b.isbn, b.cover_url, b.cover_source, b.version,
			a.id AS author_id, a.name AS author_name, a.birth_date AS author_birth_date, a.version AS author_version`).
		Joins("JOIN author_book ab ON ab.book_id = b.id").
		Joins("JOIN author a ON a.id = ab.author_id").
		Where("b.id = ?", int64(id)).
		Order("a.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: 書籍が見つかりません", common.ErrNotFound)
	}
	first := rows[0]
	d := &domainbook.BookDetail{
		ID:          domainbook.ID(first.ID),
		Title:       first.Title,
		Price:       int64(first.Price),
		Status:      first.PublishStatus,
		ISBN:        first.ISBN,
		AmazonURL:   amazonURLOf(first.ISBN),
		CoverURL:    first.CoverURL,
		CoverSource: first.CoverSource,
		Version:     first.Version,
		Authors:     make([]domainbook.AuthorSummary, 0, len(rows)),
	}
	for _, r := range rows {
		var birthDate *string
		if r.AuthorBirthDate != nil {
			s := r.AuthorBirthDate.Format(time.DateOnly)
			birthDate = &s
		}
		d.Authors = append(d.Authors, domainbook.AuthorSummary{
			ID: domainauthor.ID(r.AuthorID), Name: r.AuthorName, BirthDate: birthDate, Version: r.AuthorVersion,
		})
	}
	return d, nil
}

func (q *query) FindSummariesByAuthorID(ctx context.Context, authorID domainauthor.ID) ([]*domainbook.BookSummary, error) {
	var rows []bookModel
	err := q.db.WithContext(ctx).
		Select("book.id, book.title, book.price, book.publish_status").
		Joins("JOIN author_book ab ON ab.book_id = book.id").
		Where("ab.author_id = ?", int64(authorID)).
		Order("book.id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domainbook.BookSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, &domainbook.BookSummary{ID: domainbook.ID(r.ID), Title: r.Title, Price: int64(r.Price), Status: r.PublishStatus})
	}
	return out, nil
}

// amazonURLOf derives the product link from a stored ISBN (see domainbook.ISBN.AmazonURL).
// Stored ISBNs were validated on write, so an unparsable one just yields no link.
func amazonURLOf(isbn *string) *string {
	if isbn == nil {
		return nil
	}
	v, err := domainbook.NewISBN(*isbn)
	if err != nil {
		return nil
	}
	u, ok := v.AmazonURL()
	if !ok {
		return nil
	}
	return &u
}

type listAuthorRow struct {
	BookID     int64
	AuthorID   int64
	AuthorName string
}

// FindList は書籍を ID 順に取得範囲の分だけ取得し、その書籍の著者をまとめて1クエリで引いて各行に詰める。
func (q *query) FindList(ctx context.Context, r common.ListRange) (*domainbook.BookList, error) {
	db := q.db.WithContext(ctx)
	// Count and the range fetch are separate statements, so a concurrent insert can make them
	// disagree by a row; acceptable for a list screen, and cheaper than a snapshot tx.
	var total int64
	if err := db.Model(&bookModel{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var books []bookModel
	err := db.Select("id, title, price, publish_status, isbn, cover_url, cover_source").
		Order("id").Limit(r.Limit()).Offset(r.Offset()).
		Find(&books).Error
	if err != nil {
		return nil, err
	}
	list := &domainbook.BookList{Items: make([]*domainbook.BookListItem, 0, len(books)), Total: int(total)}
	if len(books) == 0 {
		return list, nil
	}

	ids := make([]int64, 0, len(books))
	byID := make(map[int64]*domainbook.BookListItem, len(books))
	for _, b := range books {
		item := &domainbook.BookListItem{
			ID: domainbook.ID(b.ID), Title: b.Title, Price: int64(b.Price), Status: b.PublishStatus,
			ISBN: b.ISBN, AmazonURL: amazonURLOf(b.ISBN), CoverURL: b.CoverURL, CoverSource: b.CoverSource,
			Authors: []domainbook.BookListAuthor{},
		}
		ids = append(ids, b.ID)
		byID[b.ID] = item
		list.Items = append(list.Items, item)
	}

	var authors []listAuthorRow
	err = db.Table("author_book AS ab").
		Select("ab.book_id, a.id AS author_id, a.name AS author_name").
		Joins("JOIN author a ON a.id = ab.author_id").
		Where("ab.book_id IN ?", ids).
		Order("ab.book_id, a.id").
		Scan(&authors).Error
	if err != nil {
		return nil, err
	}
	for _, a := range authors {
		item := byID[a.BookID]
		item.Authors = append(item.Authors, domainbook.BookListAuthor{ID: domainauthor.ID(a.AuthorID), Name: a.AuthorName})
	}
	return list, nil
}
