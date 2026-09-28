package tag

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

type RegisterRequest struct {
	Name string `json:"name" validate:"required,max=30" example:"データベース"`
} // @name RegisterTagRequest

type RenameRequest struct {
	Name    string `json:"name" validate:"required,max=30" example:"分散システム"`
	Version *int   `json:"version" validate:"required" example:"1"`
} // @name RenameTagRequest

type Response struct {
	ID      int64  `json:"id" example:"1"`
	Name    string `json:"name" example:"データベース"`
	Version int    `json:"version" example:"1"`
} // @name TagResponse

type ListResponse struct {
	Items []Response `json:"items"`
} // @name TagListResponse

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	RegisterUC tagcmd.RegisterUsecase
	RenameUC   tagcmd.RenameUsecase
	DeleteUC   tagcmd.DeleteUsecase
	ListUC     tagqry.ListUsecase
	// AdminOnly は追加・改名・削除に掛ける認証（閲覧は誰でもできる）。
	AdminOnly echo.MiddlewareFunc
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.RegisterTag, h.AdminOnly)
	g.PUT("/:id", h.Rename, h.AdminOnly)
	g.DELETE("/:id", h.Delete, h.AdminOnly)
}

// List godoc
// @Summary      分野タグの一覧を取得する
// @Tags         tags
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/tags [get]
func (h *Handler) List(c echo.Context) error {
	out, err := h.ListUC.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]Response, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, Response{ID: int64(it.ID), Name: it.Name})
	}
	return c.JSON(http.StatusOK, ListResponse{Items: items})
}

// RegisterTag godoc
// @Summary      分野タグを追加する（自分だけ）
// @Description  同じ名前のタグは登録できない（409）
// @Tags         tags
// @Accept       json
// @Produce      json
// @Security     AdminToken
// @Param        body body RegisterRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/tags [post]
func (h *Handler) RegisterTag(c echo.Context) error {
	req, err := common.BindValidate[RegisterRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RegisterUC.Execute(c.Request().Context(), tagcmd.RegisterCommand{Name: req.Name})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Rename godoc
// @Summary      分野タグの名前を変更する（自分だけ）
// @Description  version が一致しない場合は 409
// @Tags         tags
// @Accept       json
// @Produce      json
// @Security     AdminToken
// @Param        id   path int           true "タグのID"
// @Param        body body RenameRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/tags/{id} [put]
func (h *Handler) Rename(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	req, err := common.BindValidate[RenameRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RenameUC.Execute(c.Request().Context(), tagcmd.RenameCommand{ID: id, Name: req.Name, Version: *req.Version})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Delete godoc
// @Summary      分野タグを削除する（自分だけ）
// @Description  付いていた本からは自動で外れる
// @Tags         tags
// @Security     AdminToken
// @Param        id path int true "タグのID"
// @Success      204
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/tags/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := h.DeleteUC.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func toResponse(v *tagcmd.TagView) Response {
	return Response{ID: v.ID, Name: v.Name, Version: v.Version}
}
