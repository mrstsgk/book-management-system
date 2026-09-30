package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	domainauth "github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
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

// repository は admin_session テーブルに対して domainauth.SessionRepository を実装する。
type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainauth.SessionRepository {
	return &repository{db: db}
}

// Save は行を挿入する。
func (r *repository) Save(ctx context.Context, s *domainauth.Session) error {
	row := model{ID: string(s.ID), ExpiresAt: s.ExpiresAt, AbsoluteExpiresAt: s.AbsoluteExpiresAt, CreatedAt: time.Now()}
	return r.db.WithContext(ctx).Create(&row).Error
}

// Find は id の行を返す。無ければ common.ErrNotFound を返す。
func (r *repository) Find(ctx context.Context, id domainauth.SessionID) (*domainauth.Session, error) {
	var row model
	if err := r.db.WithContext(ctx).First(&row, "id = ?", string(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: セッションが見つかりません", domaincommon.ErrNotFound)
		}
		return nil, err
	}
	return &domainauth.Session{ID: domainauth.SessionID(row.ID), ExpiresAt: row.ExpiresAt, AbsoluteExpiresAt: row.AbsoluteExpiresAt}, nil
}

// Delete は id の行を消す。無くてもエラーにしない（ログアウトは冪等でよい）。
func (r *repository) Delete(ctx context.Context, id domainauth.SessionID) error {
	return r.db.WithContext(ctx).Delete(&model{}, "id = ?", string(id)).Error
}

// UpdateExpiry は id のアイドル期限だけを更新する。
func (r *repository) UpdateExpiry(ctx context.Context, id domainauth.SessionID, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model{}).Where("id = ?", string(id)).Update("expires_at", expiresAt).Error
}
