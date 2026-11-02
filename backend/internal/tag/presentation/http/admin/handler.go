package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
)

// Handler は、分野タグの追加・改名・削除の API を処理する（自分だけ）。
type Handler struct {
	RegisterTagUsecase tagcmd.RegisterTagUsecase
	RenameTagUsecase   tagcmd.RenameTagUsecase
	DeleteTagUsecase   tagcmd.DeleteTagUsecase
}

// RegisterTag は、分野タグを追加する（POST /api/admin/tags）。
func (h *Handler) RegisterTag(c echo.Context) error {
	var req webapi.RegisterTagRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	out, err := h.RegisterTagUsecase.Execute(c.Request().Context(), tagcmd.RegisterTagCommand{Name: req.Name})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTagResponse(out))
}

// RenameTag は、分野タグの名前を変える（PUT /api/admin/tags/{id}）。
func (h *Handler) RenameTag(c echo.Context, id int64) error {
	var req webapi.RenameTagRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	out, err := h.RenameTagUsecase.Execute(c.Request().Context(), tagcmd.RenameTagCommand{ID: id, Name: req.Name, Version: req.Version})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTagResponse(out))
}

// DeleteTag は、分野タグを削除する（DELETE /api/admin/tags/{id}）。
func (h *Handler) DeleteTag(c echo.Context, id int64) error {
	if err := h.DeleteTagUsecase.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func toTagResponse(v *tagcmd.TagView) webapi.TagResponse {
	return webapi.TagResponse{ID: v.ID, Name: v.Name, Version: v.Version}
}
