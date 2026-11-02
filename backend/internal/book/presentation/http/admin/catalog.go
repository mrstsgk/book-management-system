package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
)

// CatalogHandler は、登録前に外部カタログを確かめる API を処理する（自分だけ）。
type CatalogHandler struct {
	LookupCatalogUsecase bookqry.LookupCatalogUsecase
}

// LookupCatalog は、ISBN で外部カタログの書誌と書影を確かめる（GET /api/admin/catalog/{isbn}）。
func (h *CatalogHandler) LookupCatalog(c echo.Context, isbn string) error {
	out, err := h.LookupCatalogUsecase.Execute(c.Request().Context(), isbn)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toCatalogResponse(out))
}

func toCatalogResponse(e *bookdomain.CatalogEntry) adminapi.CatalogResponse {
	b := e.Bibliography
	res := adminapi.CatalogResponse{
		Isbn: e.ISBN.String(), Title: b.Title(), Authors: b.Authors(), Publisher: b.Publisher(), PublishedOn: b.PublishedOn(),
	}
	if u, ok := e.ISBN.AmazonURL(); ok {
		res.AmazonURL = &u
	}
	if e.Cover != nil {
		u, src := e.Cover.URL(), adminapi.CatalogResponseCoverSource(e.Cover.Source())
		res.CoverURL, res.CoverSource = &u, &src
	}
	return res
}
