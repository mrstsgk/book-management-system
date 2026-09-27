package author

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type model struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string     `gorm:"column:name;size:100;not null"`
	BirthDate *time.Time `gorm:"column:birth_date;type:date"`
	Version   int        `gorm:"column:version;not null"`
}

func (model) TableName() string {
	return "author"
}

// repository は author テーブルに対して domainauthor.Repository を実装する。
type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainauthor.Repository {
	return &repository{db: db}
}

// FindByID は id の著者行を取得する。存在しなければ common.ErrNotFound を返す。
func (r *repository) FindByID(ctx context.Context, id domainauthor.ID) (*domainauthor.Author, error) {
	var row model
	if err := r.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 著者が見つかりません", common.ErrNotFound)
		}
		return nil, err
	}
	return adapt(row)
}

// CountByIDs は ids のうち存在する著者行の件数を返す。
func (r *repository) CountByIDs(ctx context.Context, ids []domainauthor.ID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	if err := r.db.WithContext(ctx).Model(&model{}).Where("id IN ?", toInt64s(ids)).Count(&n).Error; err != nil {
		return 0, err
	}
	return int(n), nil
}

// Create は著者行を初期バージョン 1 で挿入し、採番された ID とバージョンを a に設定する。
func (r *repository) Create(ctx context.Context, a *domainauthor.Author) error {
	row := model{Name: a.Name.String(), BirthDate: birthDateTime(a.BirthDate), Version: 1}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	a.ID = domainauthor.ID(row.ID)
	a.Version = row.Version
	return nil
}

// Update は ID とバージョンが一致する著者行を更新してバージョンを進める。一致しなければ common.ErrConflict を返す。
func (r *repository) Update(ctx context.Context, a *domainauthor.Author) error {
	next := a.Version + 1
	res := r.db.WithContext(ctx).Model(&model{}).
		Where("id = ? AND version = ?", int64(a.ID), a.Version).
		Updates(map[string]any{"name": a.Name.String(), "birth_date": birthDateTime(a.BirthDate), "version": next})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("%w: 著者の更新に失敗しました（他の更新と競合しました）", common.ErrConflict)
	}
	a.Version = next
	return nil
}

func adapt(row model) (*domainauthor.Author, error) {
	name, err := domainauthor.NewName(row.Name)
	if err != nil {
		return nil, err
	}
	a := &domainauthor.Author{ID: domainauthor.ID(row.ID), Name: name, Version: row.Version}
	if row.BirthDate != nil {
		b := domainauthor.RestoreBirthDate(*row.BirthDate)
		a.BirthDate = &b
	}
	return a, nil
}

func birthDateTime(b *domainauthor.BirthDate) *time.Time {
	if b == nil {
		return nil
	}
	t := b.Time()
	return &t
}

func toInt64s(ids []domainauthor.ID) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		out = append(out, int64(id))
	}
	return out
}
