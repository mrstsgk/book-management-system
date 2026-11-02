package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

type model struct {
	ID                string    `gorm:"column:id;primaryKey;size:64"`
	ExpiresAt         time.Time `gorm:"column:expires_at;not null"`
	AbsoluteExpiresAt time.Time `gorm:"column:absolute_expires_at;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;not null"`
}

func (model) TableName() string {
	return "admin_session"
}

// repository は admin_session テーブルに対して authdomain.SessionRepository を実装する。
type repository struct {
	db *gorm.DB
}

// NewRepository は、db を使う SessionRepository を作る。
func NewRepository(db *gorm.DB) authdomain.SessionRepository {
	return &repository{db: db}
}

// Create は行を挿入する。
func (r *repository) Create(ctx context.Context, s *authdomain.Session) error {
	row := model{ID: string(s.ID), ExpiresAt: s.ExpiresAt, AbsoluteExpiresAt: s.AbsoluteExpiresAt, CreatedAt: time.Now()}
	return r.db.WithContext(ctx).Create(&row).Error
}

// FindByID は id の行を返す。無ければ common.ErrNotFound を返す。
func (r *repository) FindByID(ctx context.Context, id authdomain.SessionID) (*authdomain.Session, error) {
	var row model
	if err := r.db.WithContext(ctx).First(&row, "id = ?", string(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: セッションが見つかりません", domainerr.ErrNotFound)
		}
		return nil, err
	}
	return &authdomain.Session{ID: authdomain.SessionID(row.ID), ExpiresAt: row.ExpiresAt, AbsoluteExpiresAt: row.AbsoluteExpiresAt}, nil
}

// Delete は id の行を消す。無くてもエラーにしない（ログアウトは冪等でよい）。
func (r *repository) Delete(ctx context.Context, id authdomain.SessionID) error {
	return r.db.WithContext(ctx).Delete(&model{}, "id = ?", string(id)).Error
}

// UpdateExpiry は id のアイドル期限だけを更新する。
func (r *repository) UpdateExpiry(ctx context.Context, id authdomain.SessionID, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model{}).Where("id = ?", string(id)).Update("expires_at", expiresAt).Error
}
