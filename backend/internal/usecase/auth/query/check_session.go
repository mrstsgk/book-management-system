package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type CheckSessionUsecase interface {
	// Execute は sessionID が有効なら nil を返してアイドル期限を延ばす。無効なら common.ErrUnauthorized。
	Execute(ctx context.Context, sessionID string) error
}

type CheckSessionUsecaseImpl struct {
	Sessions auth.SessionRepository
	Now      func() time.Time
}

// Execute は見つからない・期限切れを同じ 401 にし、有効なときだけ延長する。
func (u *CheckSessionUsecaseImpl) Execute(ctx context.Context, sessionID string) error {
	s, err := u.Sessions.Find(ctx, auth.SessionID(sessionID))
	if errors.Is(err, common.ErrNotFound) {
		return fmt.Errorf("%w: ログインしてください", common.ErrUnauthorized)
	}
	if err != nil {
		return err
	}
	now := u.Now()
	if !s.IsValid(now) {
		return fmt.Errorf("%w: ログインの期限が切れました", common.ErrUnauthorized)
	}
	s.Extend(now)
	return u.Sessions.UpdateExpiry(ctx, s.ID, s.ExpiresAt)
}
