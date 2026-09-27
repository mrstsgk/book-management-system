package book

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// yen maps the NUMERIC(10,2) price column to whole yen. The pgx driver returns NUMERIC
// as a decimal string ("1500.00"), which database/sql cannot scan into int64 directly.
type yen int64

// Scan は NUMERIC の文字列表現を整数の円に変換する。小数部が 0 以外なら端数を黙って捨てずにエラーにする。
func (y *yen) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case int64:
		*y = yen(v)
		return nil
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported price type %T", src)
	}
	whole, frac, _ := strings.Cut(s, ".")
	if strings.Trim(frac, "0") != "" {
		return fmt.Errorf("price %q has a fractional part", s)
	}
	n, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return fmt.Errorf("price %q: %w", s, err)
	}
	*y = yen(n)
	return nil
}

func (y yen) Value() (driver.Value, error) {
	return int64(y), nil
}

type bookModel struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Title         string `gorm:"column:title;size:255;not null"`
	Price         yen    `gorm:"column:price;type:numeric(10,2);not null"`
	PublishStatus int    `gorm:"column:publish_status;not null"`
	Version       int    `gorm:"column:version;not null"`
}

func (bookModel) TableName() string {
	return "book"
}

type authorBookModel struct {
	AuthorID int64 `gorm:"column:author_id;primaryKey"`
	BookID   int64 `gorm:"column:book_id;primaryKey"`
	Version  int   `gorm:"column:version;not null"`
}

func (authorBookModel) TableName() string {
	return "author_book"
}

// repository は book / author_book テーブルに対して domainbook.Repository を実装する。
type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainbook.Repository {
	return &repository{db: db}
}

// FindByID は id の書籍行と紐づく著者IDを取得する。存在しなければ common.ErrNotFound を返す。
func (r *repository) FindByID(ctx context.Context, id domainbook.ID) (*domainbook.Book, error) {
	db := r.db.WithContext(ctx)
	var row bookModel
	if err := db.First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 書籍が見つかりません", common.ErrNotFound)
		}
		return nil, err
	}
	var authorIDs []int64
	if err := db.Model(&authorBookModel{}).Where("book_id = ?", row.ID).Order("author_id").Pluck("author_id", &authorIDs).Error; err != nil {
		return nil, err
	}
	return adapt(row, authorIDs)
}

// Create は書籍行を初期バージョン 1 で挿入し、著者との関連行とあわせて同一トランザクションで保存する。
func (r *repository) Create(ctx context.Context, b *domainbook.Book) error {
	row := bookModel{Title: b.Title.String(), Price: yen(b.Price.Int64()), PublishStatus: int(b.Status), Version: 1}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return insertAuthorBooks(tx, row.ID, b.AuthorIDs)
	})
	if err != nil {
		return err
	}
	b.ID = domainbook.ID(row.ID)
	b.Version = row.Version
	return nil
}

// Update は ID とバージョンが一致する書籍行を更新し、著者との関連行を入れ替える。一致しなければ common.ErrConflict を返す。
func (r *repository) Update(ctx context.Context, b *domainbook.Book) error {
	next := b.Version + 1
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&bookModel{}).
			Where("id = ? AND version = ?", int64(b.ID), b.Version).
			Updates(map[string]any{"title": b.Title.String(), "price": b.Price.Int64(), "publish_status": int(b.Status), "version": next})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("%w: 書籍の更新に失敗しました（他の更新と競合しました）", common.ErrConflict)
		}
		if err := tx.Where("book_id = ?", int64(b.ID)).Delete(&authorBookModel{}).Error; err != nil {
			return err
		}
		return insertAuthorBooks(tx, int64(b.ID), b.AuthorIDs)
	})
	if err != nil {
		return err
	}
	b.Version = next
	return nil
}

func insertAuthorBooks(tx *gorm.DB, bookID int64, authorIDs []domainauthor.ID) error {
	rows := make([]authorBookModel, 0, len(authorIDs))
	for _, id := range authorIDs {
		rows = append(rows, authorBookModel{AuthorID: int64(id), BookID: bookID, Version: 1})
	}
	return tx.Create(&rows).Error
}

func adapt(row bookModel, authorIDs []int64) (*domainbook.Book, error) {
	title, err := domainbook.NewTitle(row.Title)
	if err != nil {
		return nil, err
	}
	price, err := domainbook.NewPrice(int64(row.Price))
	if err != nil {
		return nil, err
	}
	status, err := domainbook.NewPublishStatus(row.PublishStatus)
	if err != nil {
		return nil, err
	}
	ids := make([]domainauthor.ID, 0, len(authorIDs))
	for _, id := range authorIDs {
		ids = append(ids, domainauthor.ID(id))
	}
	b, err := domainbook.New(title, price, ids, status)
	if err != nil {
		// A book without authors breaks the invariant enforced on every write: data corruption.
		return nil, fmt.Errorf("book %d is inconsistent: %w", row.ID, err)
	}
	b.ID = domainbook.ID(row.ID)
	b.Version = row.Version
	return b, nil
}
