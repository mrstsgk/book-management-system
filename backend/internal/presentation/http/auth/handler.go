package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

type LoginRequest struct {
	ID       string `json:"id" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"correct horse battery staple"`
} // @name LoginRequest

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	LoginUsecase        authcmd.LoginUsecase
	LogoutUsecase       authcmd.LogoutUsecase
	CheckSessionUsecase authqry.CheckSessionUsecase
	// SameOrigin は状態を変える要求（login / logout）に掛ける CSRF 対策。
	SameOrigin echo.MiddlewareFunc
}

// Register は、ログインの API のルートを g に登録する。
func (h *Handler) Register(g *echo.Group) {
	g.POST("/login", h.Login, h.SameOrigin)
	g.POST("/logout", h.Logout, h.SameOrigin)
	g.GET("/session", h.Session, httpserver.RequireAdminSession(h.CheckSessionUsecase))
}

// Login は、ID とパスワードでログインしてセッション Cookie を発行する（POST /api/auth/login）。
// @Summary      管理画面にログインする
// @Description  ID とパスワードが合えばセッションを発行し、httpOnly Cookie（admin_session）で返す。どちらが違うかは返さない。失敗が続くとしばらく 429
// @Tags         auth
// @Accept       json
// @Param        body body LoginRequest true "body"
// @Success      204
// @Failure      400 {object} httpserver.ErrorResponse
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      429 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	req, err := httpserver.BindValidate[LoginRequest](c)
	if err != nil {
		return err
	}
	id, err := h.LoginUsecase.Execute(c.Request().Context(), authcmd.LoginCommand{ID: req.ID, Password: req.Password})
	if err != nil {
		return err
	}
	httpserver.SetSessionCookie(c, id)
	return c.NoContent(http.StatusNoContent)
}

// Logout は、セッションを破棄して Cookie を消す（POST /api/auth/logout）。
// @Summary      ログアウトする
// @Description  サーバー側のセッションを消し、Cookie も消す。ログインしていなくても 204
// @Tags         auth
// @Success      204
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	if ck, err := c.Cookie(httpserver.SessionCookieName); err == nil && ck.Value != "" {
		if err := h.LogoutUsecase.Execute(c.Request().Context(), ck.Value); err != nil {
			return err
		}
	}
	httpserver.ClearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// Session は、ログイン中かを確かめる（GET /api/auth/session）。
// @Summary      ログイン中かを確かめる
// @Description  ログイン済みの Cookie（admin_session）が有効なら 204。画面のガードに使う。有効ならアイドル期限が延びる
// @Tags         auth
// @Success      204
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/auth/session [get]
func (h *Handler) Session(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
