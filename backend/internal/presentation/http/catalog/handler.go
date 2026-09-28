package catalog

import (
	"net/http"

	"github.com/labstack/echo/v4"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// isbnMaxLength matches the books API's isbn field (13 digits plus hyphens).
const isbnMaxLength = 17

type Response struct {
	ISBN  string `json:"isbn" example:"9784873118703"`
	Title string `json:"title" example:"データ指向アプリケーションデザイン"`
	// Authors is the provider's raw text and may include translators; the user picks the authors.
	Authors     string  `json:"authors" example:"Kleppmann,Martin 斉藤,太郎 玉川,竜司"`
	Publisher   string  `json:"publisher" example:"オライリー・ジャパン"`
	PublishedOn string  `json:"publishedOn" example:"201907"`
	Price       *int64  `json:"price" example:"4600"`
	AmazonURL   *string `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
} // @name CatalogResponse

// Handler converts HTTP ↔ UseCase only (no business logic).
type Handler struct {
	LookupUC bookqry.LookupCatalogUsecase
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/:isbn", h.Lookup)
}

// Lookup godoc
// @Summary      ISBNから書誌情報と書影を取得する
// @Description  書籍登録の入力補助。openBD を優先し、書影が無ければ楽天ブックスで補う。著者は提供元の文字列のまま返す
// @Tags         catalog
// @Produce      json
// @Param        isbn path string true "ISBN（13桁または10桁。ハイフン可）"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/catalog/{isbn} [get]
func (h *Handler) Lookup(c echo.Context) error {
	isbn := c.Param("isbn")
	if len(isbn) > isbnMaxLength {
		return &common.ValidationError{Fields: []common.FieldError{{Field: "isbn", Rule: "max"}}}
	}
	out, err := h.LookupUC.Execute(c.Request().Context(), isbn)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

func toResponse(e *domainbook.CatalogEntry) Response {
	res := Response{
		ISBN: e.ISBN.String(), Title: e.Title, Authors: e.Authors, Publisher: e.Publisher,
		PublishedOn: e.PublishedOn, Price: e.Price,
	}
	if u, ok := e.ISBN.AmazonURL(); ok {
		res.AmazonURL = &u
	}
	if e.Cover != nil {
		u, src := e.Cover.URL(), string(e.Cover.Source())
		res.CoverURL, res.CoverSource = &u, &src
	}
	return res
}
