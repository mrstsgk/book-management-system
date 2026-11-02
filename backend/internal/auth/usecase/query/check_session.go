package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// CheckSessionUsecase は、セッションが有効かを確かめ、有効なら期限を延ばす。
type CheckSessionUsecase interface {
	// Execute は sessionID が有効なら nil を返してアイドル期限を延ばす。無効なら domainerr.ErrUnauthorized。
	Execute(ctx context.Context, sessionID string) error
}

type CheckSessionUsecaseImpl struct {
	Sessions authdomain.SessionRepository
	Now      func() time.Time
}

// Execute は見つからない・期限切れを同じ 401 にし、有効なときだけ延長する。
func (u *CheckSessionUsecaseImpl) Execute(ctx context.Context, sessionID string) error {
	s, err := u.Sessions.FindByID(ctx, authdomain.SessionID(sessionID))
	if errors.Is(err, domainerr.ErrNotFound) {
		return fmt.Errorf("%w: ログインしてください", domainerr.ErrUnauthorized)
	}
	if err != nil {
		return err
	}
	now := u.Now()
	if !s.IsValid(now) {
		return fmt.Errorf("%w: ログインの期限が切れました", domainerr.ErrUnauthorized)
	}
	s.Extend(now)
	return u.Sessions.UpdateExpiry(ctx, s.ID, s.ExpiresAt)
}
