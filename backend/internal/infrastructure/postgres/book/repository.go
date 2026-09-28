package book

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type model struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ISBN          string    `gorm:"column:isbn;size:13;not null"`
	Title         string    `gorm:"column:title;size:255;not null"`
	TitleOverride *string   `gorm:"column:title_override;size:255"`
	Authors       string    `gorm:"column:authors;size:500;not null"`
	Publisher     string    `gorm:"column:publisher;size:255;not null"`
	PublishedOn   string    `gorm:"column:published_on;size:32;not null"`
	CoverURL      *string   `gorm:"column:cover_url;size:2048"`
	CoverSource   *string   `gorm:"column:cover_source;size:16"`
	Summary       string    `gorm:"column:summary;size:100;not null"`
	Comment       string    `gorm:"column:comment;not null"`
	Rating        int       `gorm:"column:rating;not null"`
	Version       int       `gorm:"column:version;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

func (model) TableName() string {
	return "book"
}

// repository は book テーブルに対して domainbook.Repository を実装する。
type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainbook.Repository {
	return &repository{db: db}
}

// FindByID は id の行を取得する。存在しなければ common.ErrNotFound を返す。
func (r *repository) FindByID(ctx context.Context, id domainbook.ID) (*domainbook.Book, error) {
	var row model
	if err := r.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 読んだ本が見つかりません", common.ErrNotFound)
		}
		return nil, err
	}
	return adapt(row)
}

// Create は行を初期バージョン 1 で挿入し、採番した ID とバージョンを b に設定する。同じ ISBN があれば common.ErrConflict を返す。
func (r *repository) Create(ctx context.Context, b *domainbook.Book) error {
	row := toModel(b)
	row.Version = 1
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return translateDuplicateISBN(err)
	}
	b.ID = domainbook.ID(row.ID)
	b.Version = row.Version
	return nil
}

// Update は ID とバージョンが一致する行を更新してバージョンを進める。一致しなければ common.ErrConflict を返す。
func (r *repository) Update(ctx context.Context, b *domainbook.Book) error {
	row := toModel(b)
	next := b.Version + 1
	res := r.db.WithContext(ctx).Model(&model{}).
		Where("id = ? AND version = ?", int64(b.ID), b.Version).
		Updates(map[string]any{
			"title": row.Title, "title_override": row.TitleOverride, "authors": row.Authors,
			"publisher": row.Publisher, "published_on": row.PublishedOn,
			"cover_url": row.CoverURL, "cover_source": row.CoverSource,
			"summary": row.Summary, "comment": row.Comment, "rating": row.Rating,
			"version": next, "updated_at": gorm.Expr("NOW()"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("%w: 読んだ本の更新に失敗しました（他の更新と競合しました）", common.ErrConflict)
	}
	b.Version = next
	return nil
}

// Delete は id の行を削除する。存在しなければ common.ErrNotFound を返す。
func (r *repository) Delete(ctx context.Context, id domainbook.ID) error {
	res := r.db.WithContext(ctx).Delete(&model{}, int64(id))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("%w: 読んだ本が見つかりません", common.ErrNotFound)
	}
	return nil
}

// uniqueViolation は PostgreSQL の一意制約違反の SQLSTATE。
const uniqueViolation = "23505"

// translateDuplicateISBN は ISBN の一意制約違反を ErrConflict にする（同じ本の二重登録を 500 ではなく 409 にするため）。
func translateDuplicateISBN(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "uq_book_isbn" {
		return fmt.Errorf("%w: 同じISBNの本が既に登録されています", common.ErrConflict)
	}
	return err
}

func toModel(b *domainbook.Book) model {
	row := model{
		ID:          int64(b.ID),
		ISBN:        b.ISBN.String(),
		Title:       b.Bibliography.Title(),
		Authors:     b.Bibliography.Authors(),
		Publisher:   b.Bibliography.Publisher(),
		PublishedOn: b.Bibliography.PublishedOn(),
		Summary:     b.Summary.String(),
		Comment:     b.Comment.String(),
		Rating:      b.Rating.Int(),
		Version:     b.Version,
	}
	if b.Cover != nil {
		u, s := b.Cover.URL(), string(b.Cover.Source())
		row.CoverURL, row.CoverSource = &u, &s
	}
	if b.TitleOverride != nil {
		v := b.TitleOverride.String()
		row.TitleOverride = &v
	}
	return row
}

// adapt は行を VO で検証し直して Book に戻す。書き込み時に検証済みの値なので、失敗はデータの破損を意味する。
func adapt(row model) (*domainbook.Book, error) {
	isbn, err := domainbook.NewISBN(row.ISBN)
	if err != nil {
		return nil, err
	}
	bib, err := domainbook.NewBibliography(row.Title, row.Authors, row.Publisher, row.PublishedOn)
	if err != nil {
		return nil, err
	}
	summary, err := domainbook.NewSummary(row.Summary)
	if err != nil {
		return nil, err
	}
	comment, err := domainbook.NewComment(row.Comment)
	if err != nil {
		return nil, err
	}
	rating, err := domainbook.NewRating(row.Rating)
	if err != nil {
		return nil, err
	}
	var cover *domainbook.Cover
	if row.CoverURL != nil && row.CoverSource != nil {
		c, err := domainbook.NewCover(*row.CoverURL, domainbook.CoverSource(*row.CoverSource))
		if err != nil {
			return nil, err
		}
		cover = &c
	}
	override, err := adaptTitleOverride(row.TitleOverride)
	if err != nil {
		return nil, err
	}
	b := domainbook.New(isbn, bib, cover, summary, comment, rating)
	b.OverrideTitle(override)
	b.ID = domainbook.ID(row.ID)
	b.Version = row.Version
	return b, nil
}

// adaptTitleOverride は保存済みの上書きを VO で検証し直す。NULL なら上書きなし。
func adaptTitleOverride(v *string) (*domainbook.Title, error) {
	if v == nil {
		return nil, nil
	}
	t, err := domainbook.NewTitle(*v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
