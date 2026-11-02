package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
)

// Handler は、読んだ本の登録・更新・削除の API を処理する（自分だけ）。
type Handler struct {
	RegisterBookUsecase bookcmd.RegisterBookUsecase
	UpdateBookUsecase   bookcmd.UpdateBookUsecase
	DeleteBookUsecase   bookcmd.DeleteBookUsecase
}

// RegisterBook は、読んだ本を登録する（POST /api/admin/books）。
func (h *Handler) RegisterBook(c echo.Context) error {
	var req webapi.RegisterBookRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	out, err := h.RegisterBookUsecase.Execute(c.Request().Context(), bookcmd.RegisterBookCommand{
		ISBN: req.Isbn, Summary: req.Summary, TitleOverride: deref(req.TitleOverride), TagIDs: derefSlice(req.TagIds), Comment: req.Comment, Rating: req.Rating,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bookpublic.ToBookResponse(out))
}

// UpdateBook は、読んだ本の一言まとめ・感想・評価・書名の上書き・分野タグを更新する（PUT /api/admin/books/{id}）。
func (h *Handler) UpdateBook(c echo.Context, id int64) error {
	var req webapi.UpdateBookRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	out, err := h.UpdateBookUsecase.Execute(c.Request().Context(), bookcmd.UpdateBookCommand{
		ID: id, Summary: req.Summary, TitleOverride: deref(req.TitleOverride), TagIDs: derefSlice(req.TagIds), Comment: req.Comment, Rating: req.Rating, Version: req.Version,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bookpublic.ToBookResponse(out))
}

// DeleteBook は、読んだ本を削除する（DELETE /api/admin/books/{id}）。
func (h *Handler) DeleteBook(c echo.Context, id int64) error {
	if err := h.DeleteBookUsecase.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// 省略された任意の項目は、UseCase には空として渡す（空の書名の上書き・空のタグは「付けない」と同じ意味のため）。
func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefSlice(p *[]int64) []int64 {
	if p == nil {
		return nil
	}
	return *p
}
