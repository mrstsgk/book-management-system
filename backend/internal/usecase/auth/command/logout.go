package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
)

type LogoutUsecase interface {
	Execute(ctx context.Context, sessionID string) error
}

type LogoutUsecaseImpl struct {
	Sessions auth.SessionRepository
}

// Execute はセッションを消す。既に無くてもエラーにしない。
func (u *LogoutUsecaseImpl) Execute(ctx context.Context, sessionID string) error {
	return u.Sessions.Delete(ctx, auth.SessionID(sessionID))
}
