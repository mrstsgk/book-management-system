package public

import (
	"net/http"

	"github.com/labstack/echo/v4"

	// swag が注釈の httpserver.ErrorResponse を解決するために読む（コードからは使わない）
	_ "github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

type Response struct {
	ID      int64  `json:"id" example:"1"`
	Name    string `json:"name" example:"データベース"`
	Version int    `json:"version" example:"1"`
} // @name TagResponse

type ListResponse struct {
	Items []Response `json:"items"`
} // @name TagListResponse

type BookCountResponse struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"データベース"`
	BookCount int    `json:"bookCount" example:"3"`
} // @name TagBookCountResponse

type BookCountListResponse struct {
	Items []BookCountResponse `json:"items"`
} // @name TagBookCountListResponse

// Handler は、分野タグの閲覧の API を処理する（誰でも使える）。
type Handler struct {
	SearchTagsUsecase tagqry.SearchTagsUsecase
	CountBooksUsecase tagqry.CountBooksUsecase
}

// Register は、分野タグの閲覧の API のルートを g に登録する。
func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.Search)
	g.GET("/counts", h.CountBooks)
}

// Search は、分野タグの一覧を参照する（GET /api/tags）。
// @Summary      分野タグの一覧を取得する
// @Tags         tags
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/tags [get]
func (h *Handler) Search(c echo.Context) error {
	out, err := h.SearchTagsUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]Response, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, Response{ID: int64(it.ID), Name: it.Name, Version: it.Version})
	}
	return c.JSON(http.StatusOK, ListResponse{Items: items})
}

// CountBooks は、分野タグごとの冊数を返す（GET /api/tags/counts）。
// @Summary      分野タグごとの冊数を取得する
// @Description  本が1冊も付いていないタグは含めない。冊数の多い順、同数ならタグ名順
// @Tags         tags
// @Produce      json
// @Success      200 {object} BookCountListResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/tags/counts [get]
func (h *Handler) CountBooks(c echo.Context) error {
	out, err := h.CountBooksUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]BookCountResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, BookCountResponse{ID: int64(it.ID), Name: it.Name, BookCount: it.BookCount})
	}
	return c.JSON(http.StatusOK, BookCountListResponse{Items: items})
}
