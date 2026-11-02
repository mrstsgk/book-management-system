package public

import (
	"net/http"

	"github.com/labstack/echo/v4"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
)

// Handler は、読んだ本の閲覧の API を処理する（誰でも使える）。
type Handler struct {
	ReadBookUsecase    bookqry.ReadBookUsecase
	SearchBooksUsecase bookqry.SearchBooksUsecase
}

// SearchBooks は、読んだ本を検索する（GET /api/books）。
func (h *Handler) SearchBooks(c echo.Context, params webapi.SearchBooksParams) error {
	// 既定値は spec にも書いてあるが、生成コードは既定値を入れないのでここで補う
	limit, offset := bookdomain.DefaultListLimit, 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	var keyword string
	if params.Q != nil {
		keyword = *params.Q
	}
	out, err := h.SearchBooksUsecase.Execute(c.Request().Context(), bookqry.SearchBooksInput{Keyword: keyword, TagID: params.TagID, Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	items := make([]webapi.BookListItemResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, webapi.BookListItemResponse{
			ID: int64(it.ID), Isbn: it.ISBN, Title: it.Title, Summary: it.Summary, Tags: it.Tags, Authors: it.Authors,
			AmazonURL: it.AmazonURL, CoverURL: it.CoverURL, CoverSource: (*webapi.BookListItemResponseCoverSource)(it.CoverSource), Rating: it.Rating,
		})
	}
	return c.JSON(http.StatusOK, webapi.SearchBooksResponse{Items: items, Total: out.Total, Limit: limit, Offset: offset})
}

// ReadBook は、読んだ本を 1 件参照する（GET /api/books/{id}）。
func (h *Handler) ReadBook(c echo.Context, id int64) error {
	out, err := h.ReadBookUsecase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ToBookResponse(out))
}

// ToBookResponse は、読んだ本の詳細をレスポンスの形にする（管理の登録・更新の応答も同じ形）。
func ToBookResponse(d *bookdomain.BookDetail) webapi.BookResponse {
	return webapi.BookResponse{
		ID: int64(d.ID), Isbn: d.ISBN, Title: d.Title, Authors: d.Authors, Publisher: d.Publisher,
		PublishedOn: d.PublishedOn, AmazonURL: d.AmazonURL, CoverURL: d.CoverURL, CoverSource: (*webapi.BookResponseCoverSource)(d.CoverSource),
		TitleOverride: d.TitleOverride, Summary: d.Summary, Tags: d.Tags, Comment: d.Comment, Rating: d.Rating, Version: d.Version,
	}
}
