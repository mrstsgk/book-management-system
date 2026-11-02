package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
)

// 閲覧（public）と同じ UseCase を呼ぶが、Handler は共有しない。チャネルごとに生成型が別のため。

// SearchTags は、分野タグの一覧を参照する（GET /api/admin/tags）。
func (h *Handler) SearchTags(c echo.Context) error {
	out, err := h.SearchTagsUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]adminapi.TagResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, adminapi.TagResponse{ID: int64(it.ID), Name: it.Name, Version: it.Version})
	}
	return c.JSON(http.StatusOK, adminapi.SearchTagsResponse{Items: items})
}

// CountBooks は、分野タグごとの冊数を返す（GET /api/admin/tags/counts）。
func (h *Handler) CountBooks(c echo.Context) error {
	out, err := h.CountBooksUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]adminapi.TagBookCountResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, adminapi.TagBookCountResponse{ID: int64(it.ID), Name: it.Name, BookCount: it.BookCount})
	}
	return c.JSON(http.StatusOK, adminapi.CountBooksResponse{Items: items})
}
