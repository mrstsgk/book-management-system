package httpserver

import (
	"fmt"

	"github.com/labstack/echo/v4"

	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// RequireAdminSession は書き込み系の API を自分だけが使えるようにする。Cookie のセッションが有効でなければ 401。
// 有効なら UseCase 側でアイドル期限が延びる。
func RequireAdminSession(check authqry.CheckSessionUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ck, err := c.Cookie(SessionCookieName)
			if err != nil || ck.Value == "" {
				return fmt.Errorf("%w: ログインしてください", domainerr.ErrUnauthorized)
			}
			if err := check.Execute(c.Request().Context(), ck.Value); err != nil {
				return err
			}
			return next(c)
		}
	}
}
