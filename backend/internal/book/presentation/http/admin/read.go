package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
)

// 閲覧（public）と同じ UseCase を呼ぶが、Handler は共有しない。チャネルごとに生成型が別で、
// 管理の画面だけの項目（版など）を足すときに閲覧の API へ影響させないため。

// SearchBooks は、読んだ本を検索する（GET /api/admin/books）。
func (h *Handler) SearchBooks(c echo.Context, params adminapi.SearchBooksParams) error {
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
	items := make([]adminapi.BookListItemResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, adminapi.BookListItemResponse{
			ID: int64(it.ID), Isbn: it.ISBN, Title: it.Title, Summary: it.Summary, Tags: it.Tags, Authors: it.Authors,
			AmazonURL: it.AmazonURL, CoverURL: it.CoverURL, CoverSource: (*adminapi.BookListItemResponseCoverSource)(it.CoverSource), Rating: it.Rating,
		})
	}
	return c.JSON(http.StatusOK, adminapi.SearchBooksResponse{Items: items, Total: out.Total, Limit: limit, Offset: offset})
}

// ReadBook は、読んだ本を 1 件参照する（GET /api/admin/books/{id}）。
func (h *Handler) ReadBook(c echo.Context, id int64) error {
	out, err := h.ReadBookUsecase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBookResponse(out))
}

func toBookResponse(d *bookdomain.BookDetail) adminapi.BookResponse {
	return adminapi.BookResponse{
		ID: int64(d.ID), Isbn: d.ISBN, Title: d.Title, Authors: d.Authors, Publisher: d.Publisher,
		PublishedOn: d.PublishedOn, AmazonURL: d.AmazonURL, CoverURL: d.CoverURL, CoverSource: (*adminapi.BookResponseCoverSource)(d.CoverSource),
		TitleOverride: d.TitleOverride, Summary: d.Summary, Tags: d.Tags, Comment: d.Comment, Rating: d.Rating, Version: d.Version,
	}
}
