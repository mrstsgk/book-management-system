package author

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	authorcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

type CreateRequest struct {
	Name      string  `json:"name" validate:"required,max=100" example:"太宰治"`
	BirthDate *string `json:"birthDate" validate:"omitempty,datetime=2006-01-02" example:"1909-06-19"`
} // @name CreateAuthorRequest

type UpdateRequest struct {
	Name      string  `json:"name" validate:"required,max=100" example:"太宰治"`
	BirthDate *string `json:"birthDate" validate:"omitempty,datetime=2006-01-02" example:"1909-06-19"`
	Version   *int    `json:"version" validate:"required" example:"1"`
} // @name UpdateAuthorRequest

type Response struct {
	ID        int64   `json:"id" example:"1"`
	Name      string  `json:"name" example:"太宰治"`
	BirthDate *string `json:"birthDate" example:"1909-06-19"`
	Version   int     `json:"version" example:"1"`
} // @name AuthorResponse

type BookResponse struct {
	ID     int64  `json:"id" example:"1"`
	Title  string `json:"title" example:"人間失格"`
	Price  int64  `json:"price" example:"1500"`
	Status int    `json:"status" enums:"1,2" example:"1"`
} // @name AuthorBookResponse

// Handler converts HTTP ↔ UseCase only (no business logic).
type Handler struct {
	CreateUC authorcmd.CreateUsecase
	UpdateUC authorcmd.UpdateUsecase
	BooksUC  bookqry.ListByAuthorUsecase
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.GET("/:id/books", h.ListBooks)
}

// Create godoc
// @Summary      著者を作成する
// @Description  生年月日は現在より過去日付
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        body body CreateRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/authors [post]
func (h *Handler) Create(c echo.Context) error {
	req, err := common.BindValidate[CreateRequest](c)
	if err != nil {
		return err
	}
	birthDate, err := parseDate(req.BirthDate)
	if err != nil {
		return err
	}
	out, err := h.CreateUC.Execute(c.Request().Context(), authorcmd.CreateCommand{Name: req.Name, BirthDate: birthDate})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Update godoc
// @Summary      著者を更新する
// @Description  生年月日は現在より過去日付。version が一致しない場合は 409
// @Tags         authors
// @Accept       json
// @Produce      json
// @Param        id path int true "著者ID"
// @Param        body body UpdateRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/authors/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	req, err := common.BindValidate[UpdateRequest](c)
	if err != nil {
		return err
	}
	birthDate, err := parseDate(req.BirthDate)
	if err != nil {
		return err
	}
	out, err := h.UpdateUC.Execute(c.Request().Context(), authorcmd.UpdateCommand{
		ID: id, Name: req.Name, BirthDate: birthDate, Version: *req.Version,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// ListBooks godoc
// @Summary      著者に紐づく書籍一覧を取得する
// @Description  書籍が見つからない場合は空配列を返す
// @Tags         authors
// @Produce      json
// @Param        id path int true "著者ID"
// @Success      200 {array} BookResponse
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/authors/{id}/books [get]
func (h *Handler) ListBooks(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	items, err := h.BooksUC.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	res := make([]BookResponse, 0, len(items))
	for _, b := range items {
		res = append(res, toBookResponse(b))
	}
	return c.JSON(http.StatusOK, res)
}

// parseDate only converts; the format is already checked by the validate tag and
// the "past only" rule belongs to the domain.
func parseDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := time.Parse(time.DateOnly, *s)
	if err != nil {
		return nil, &common.ValidationError{Fields: []common.FieldError{{Field: "birthDate", Rule: "datetime"}}}
	}
	return &t, nil
}

func toResponse(a *domainauthor.Author) Response {
	res := Response{ID: int64(a.ID), Name: a.Name.String(), Version: a.Version}
	if a.BirthDate != nil {
		s := a.BirthDate.Time().Format(time.DateOnly)
		res.BirthDate = &s
	}
	return res
}

func toBookResponse(b *domainbook.BookSummary) BookResponse {
	return BookResponse{ID: int64(b.ID), Title: b.Title, Price: b.Price, Status: b.Status}
}
