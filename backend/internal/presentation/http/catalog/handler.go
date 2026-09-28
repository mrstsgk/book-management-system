package catalog

import (
	"net/http"

	"github.com/labstack/echo/v4"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// isbnMaxLength は書籍 API の isbn と同じ上限（13桁＋ハイフン）。
const isbnMaxLength = 17

type Response struct {
	ISBN        string  `json:"isbn" example:"9784873118703"`
	Title       string  `json:"title" example:"データ指向アプリケーションデザイン"`
	Authors     string  `json:"authors" example:"Kleppmann,Martin 斉藤,太郎 玉川,竜司"`
	Publisher   string  `json:"publisher" example:"オーム社"`
	PublishedOn string  `json:"publishedOn" example:"201907"`
	AmazonURL   *string `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
	// CoverProductURL は楽天の商品ページ（楽天の書影のときだけ）。
	CoverProductURL *string `json:"coverProductUrl" example:"https://books.rakuten.co.jp/rb/15949390/"`
} // @name CatalogResponse

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	LookupUC bookqry.LookupCatalogUsecase
	// AdminOnly は登録画面でだけ使うための認証（外部カタログへの中継を誰にでも開放しないため）。
	AdminOnly echo.MiddlewareFunc
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/:isbn", h.Lookup, h.AdminOnly)
}

// Lookup godoc
// @Summary      ISBNで外部カタログの書誌と書影を確かめる（自分だけ）
// @Description  登録前の確認用。書誌は openBD から取得し、書影が無ければ楽天ブックスで補う
// @Tags         catalog
// @Produce      json
// @Security     AdminToken
// @Param        isbn path string true "ISBN（13桁または10桁。ハイフン可）"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
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
	b := e.Bibliography
	res := Response{
		ISBN: e.ISBN.String(), Title: b.Title(), Authors: b.Authors(), Publisher: b.Publisher(), PublishedOn: b.PublishedOn(),
	}
	if u, ok := e.ISBN.AmazonURL(); ok {
		res.AmazonURL = &u
	}
	if e.Cover != nil {
		u, src := e.Cover.URL(), string(e.Cover.Source())
		res.CoverURL, res.CoverSource = &u, &src
		if p := e.Cover.ProductURL(); p != "" {
			res.CoverProductURL = &p
		}
	}
	return res
}
