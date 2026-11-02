package command

import (
	"context"

	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
)

// LogoutUsecase は、セッションを破棄する。
type LogoutUsecase interface {
	Execute(ctx context.Context, sessionID string) error
}

type LogoutUsecaseImpl struct {
	Sessions authdomain.SessionRepository
}

// Execute はセッションを消す。既に無くてもエラーにしない。
func (u *LogoutUsecaseImpl) Execute(ctx context.Context, sessionID string) error {
	return u.Sessions.Delete(ctx, authdomain.SessionID(sessionID))
}
