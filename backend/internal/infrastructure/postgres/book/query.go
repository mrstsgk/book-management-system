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

type detailRow struct {
	ID              int64
	Title           string
	Price           yen
	PublishStatus   int
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
		Select(`b.id, b.title, b.price, b.publish_status, b.version,
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
		ID:      domainbook.ID(first.ID),
		Title:   first.Title,
		Price:   int64(first.Price),
		Status:  first.PublishStatus,
		Version: first.Version,
		Authors: make([]domainbook.AuthorSummary, 0, len(rows)),
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
