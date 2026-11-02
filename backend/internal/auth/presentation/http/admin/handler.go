package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

// Handler は、管理画面のログイン・ログアウト・ログイン中かの確認の API を処理する。
type Handler struct {
	LoginUsecase  authcmd.LoginUsecase
	LogoutUsecase authcmd.LogoutUsecase
}

// Login は、ID とパスワードでログインしてセッション Cookie を発行する（POST /api/admin/auth/login）。
func (h *Handler) Login(c echo.Context) error {
	var req webapi.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	id, err := h.LoginUsecase.Execute(c.Request().Context(), authcmd.LoginCommand{ID: req.ID, Password: req.Password})
	if err != nil {
		return err
	}
	httpserver.SetSessionCookie(c, id)
	return c.NoContent(http.StatusNoContent)
}

// Logout は、セッションを破棄して Cookie を消す（POST /api/admin/auth/logout）。
func (h *Handler) Logout(c echo.Context) error {
	if ck, err := c.Cookie(httpserver.SessionCookieName); err == nil && ck.Value != "" {
		if err := h.LogoutUsecase.Execute(c.Request().Context(), ck.Value); err != nil {
			return err
		}
	}
	httpserver.ClearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// CheckSession は、ログイン中かを確かめる（GET /api/admin/auth/session）。
// 判定とアイドル期限の延長は、この operation に掛けたセッションの middleware（RequireAdminSession）が行う。
func (h *Handler) CheckSession(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
