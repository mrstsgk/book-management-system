package public

import (
	"net/http"

	"github.com/labstack/echo/v4"

	publicapi "github.com/mrstsgk/book-management-system/backend/api/gen/public"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// Handler は、分野タグの閲覧の API を処理する（誰でも使える）。
type Handler struct {
	SearchTagsUsecase tagqry.SearchTagsUsecase
	CountBooksUsecase tagqry.CountBooksUsecase
}

// SearchTags は、分野タグの一覧を参照する（GET /api/tags）。
func (h *Handler) SearchTags(c echo.Context) error {
	out, err := h.SearchTagsUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]publicapi.TagResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, publicapi.TagResponse{ID: int64(it.ID), Name: it.Name, Version: it.Version})
	}
	return c.JSON(http.StatusOK, publicapi.SearchTagsResponse{Items: items})
}

// CountBooks は、分野タグごとの冊数を返す（GET /api/tags/counts）。
func (h *Handler) CountBooks(c echo.Context) error {
	out, err := h.CountBooksUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]publicapi.TagBookCountResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, publicapi.TagBookCountResponse{ID: int64(it.ID), Name: it.Name, BookCount: it.BookCount})
	}
	return c.JSON(http.StatusOK, publicapi.CountBooksResponse{Items: items})
}
