package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/query"
)

type LoginRequest struct {
	ID       string `json:"id" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"correct horse battery staple"`
} // @name LoginRequest

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	LoginUC  authcmd.LoginUsecase
	LogoutUC authcmd.LogoutUsecase
	CheckUC  authqry.CheckSessionUsecase
	// SameOrigin は状態を変える要求（login / logout）に掛ける CSRF 対策。
	SameOrigin echo.MiddlewareFunc
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/login", h.Login, h.SameOrigin)
	g.POST("/logout", h.Logout, h.SameOrigin)
	g.GET("/session", h.Session, common.RequireAdminSession(h.CheckUC))
}

// Login godoc
// @Summary      管理画面にログインする
// @Description  ID とパスワードが合えばセッションを発行し、httpOnly Cookie（admin_session）で返す。どちらが違うかは返さない。失敗が続くとしばらく 429
// @Tags         auth
// @Accept       json
// @Param        body body LoginRequest true "body"
// @Success      204
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      403 {object} common.ErrorResponse
// @Failure      429 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	req, err := common.BindValidate[LoginRequest](c)
	if err != nil {
		return err
	}
	id, err := h.LoginUC.Execute(c.Request().Context(), authcmd.LoginCommand{ID: req.ID, Password: req.Password})
	if err != nil {
		return err
	}
	common.SetSessionCookie(c, id)
	return c.NoContent(http.StatusNoContent)
}

// Logout godoc
// @Summary      ログアウトする
// @Description  サーバー側のセッションを消し、Cookie も消す。ログインしていなくても 204
// @Tags         auth
// @Success      204
// @Failure      403 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	if ck, err := c.Cookie(common.SessionCookieName); err == nil && ck.Value != "" {
		if err := h.LogoutUC.Execute(c.Request().Context(), ck.Value); err != nil {
			return err
		}
	}
	common.ClearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// Session godoc
// @Summary      ログイン中かを確かめる
// @Description  ログイン済みの Cookie（admin_session）が有効なら 204。画面のガードに使う。有効ならアイドル期限が延びる
// @Tags         auth
// @Success      204
// @Failure      401 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/auth/session [get]
func (h *Handler) Session(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
