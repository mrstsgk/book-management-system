package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

type model struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;size:30;not null"`
	Version   int       `gorm:"column:version;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (model) TableName() string {
	return "tags"
}

// repository は tags テーブルに対して tagdomain.Repository を実装する。
type repository struct {
	db *gorm.DB
}

// NewRepository は、db を使う Repository を作る。
func NewRepository(db *gorm.DB) tagdomain.Repository {
	return &repository{db: db}
}

// FindByID は id の行を取得する。存在しなければ common.ErrNotFound を返す。
func (r *repository) FindByID(ctx context.Context, id tagdomain.ID) (*tagdomain.Tag, error) {
	var row model
	if err := r.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.ErrNotFound, "tagNotFound", "タグが見つかりません")
		}
		return nil, err
	}
	return adapt(row)
}

// Create は行を初期バージョン 1 で挿入し、採番した ID とバージョンを t に設定する。同じ名前があれば common.ErrConflict を返す。
func (r *repository) Create(ctx context.Context, t *tagdomain.Tag) error {
	row := model{Name: t.Name.String(), Version: 1}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return translateDuplicateName(err)
	}
	t.ID = tagdomain.ID(row.ID)
	t.Version = row.Version
	return nil
}

// Update は ID とバージョンが一致する行を更新してバージョンを進める。不一致・同名なら common.ErrConflict を返す。
func (r *repository) Update(ctx context.Context, t *tagdomain.Tag) error {
	next := t.Version + 1
	res := r.db.WithContext(ctx).Model(&model{}).
		Where("id = ? AND version = ?", int64(t.ID), t.Version).
		Updates(map[string]any{"name": t.Name.String(), "version": next, "updated_at": gorm.Expr("NOW()")})
	if res.Error != nil {
		return translateDuplicateName(res.Error)
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.ErrConflict, "versionConflict", "タグの更新に失敗しました（他の更新と競合しました）")
	}
	t.Version = next
	return nil
}

// Delete は id の行を削除する。存在しなければ common.ErrNotFound を返す。付いていた本からは book_tags の
// ON DELETE CASCADE で自動的に外れる。
func (r *repository) Delete(ctx context.Context, id tagdomain.ID) error {
	res := r.db.WithContext(ctx).Delete(&model{}, int64(id))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainerr.New(domainerr.ErrNotFound, "tagNotFound", "タグが見つかりません")
	}
	return nil
}

const uniqueViolation = "23505"

// translateDuplicateName はタグ名の一意制約違反を ErrConflict にする（500 ではなく 409 にするため）。
func translateDuplicateName(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "uq_tags_name" {
		return domainerr.New(domainerr.ErrConflict, "tagNameDuplicated", "同じ名前のタグが既にあります")
	}
	return err
}

// adapt は行を VO で検証し直して Tag に戻す。書き込み時に検証済みの値なので、失敗はデータの破損を意味する。
func adapt(row model) (*tagdomain.Tag, error) {
	name, err := tagdomain.NewName(row.Name)
	if err != nil {
		return nil, err
	}
	t := tagdomain.New(name)
	t.ID = tagdomain.ID(row.ID)
	t.Version = row.Version
	return t, nil
}
