package book

import (
	"net/http"

	"github.com/labstack/echo/v4"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

type ListRequest struct {
	Limit  *int `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

type CreateRequest struct {
	Title     string  `json:"title" validate:"required,max=255" example:"人間失格"`
	Price     *int64  `json:"price" validate:"required,min=0,max=99999999" example:"1500"`
	AuthorIDs []int64 `json:"authorIds" validate:"required,min=1,dive,gt=0"`
	Status    *int    `json:"status" validate:"required,oneof=1 2" enums:"1,2" example:"1"`
	// ISBN (13 or 10 digits, hyphens allowed). The cover is looked up by it on save.
	ISBN *string `json:"isbn" validate:"omitempty,max=17" example:"9784873118703"`
} // @name CreateBookRequest

type UpdateRequest struct {
	Title     string  `json:"title" validate:"required,max=255" example:"人間失格"`
	Price     *int64  `json:"price" validate:"required,min=0,max=99999999" example:"1500"`
	AuthorIDs []int64 `json:"authorIds" validate:"required,min=1,dive,gt=0"`
	Status    *int    `json:"status" validate:"required,oneof=1 2" enums:"1,2" example:"1"`
	// Omitted or null clears the ISBN and its cover (PUT replaces the whole book).
	ISBN    *string `json:"isbn" validate:"omitempty,max=17" example:"9784873118703"`
	Version *int    `json:"version" validate:"required" example:"1"`
} // @name UpdateBookRequest

type AuthorResponse struct {
	ID        int64   `json:"id" example:"1"`
	Name      string  `json:"name" example:"太宰治"`
	BirthDate *string `json:"birthDate" example:"1909-06-19"`
	Version   int     `json:"version" example:"1"`
} // @name BookAuthorResponse

type Response struct {
	ID      int64            `json:"id" example:"1"`
	Title   string           `json:"title" example:"人間失格"`
	Price   int64            `json:"price" example:"1500"`
	Authors []AuthorResponse `json:"authors"`
	Status  int              `json:"status" enums:"1,2" example:"1"`
	ISBN    *string          `json:"isbn" example:"9784873118703"`
	// AmazonURL is derived from the ISBN; null when there is no ISBN (or no ISBN-10 form).
	AmazonURL *string `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	// CoverURL is the provider-hosted image; show it as-is with the credit for CoverSource.
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
	Version     int     `json:"version" example:"1"`
} // @name BookResponse

type ListAuthorResponse struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Martin Kleppmann"`
} // @name BookListAuthorResponse

type ListItemResponse struct {
	ID      int64                `json:"id" example:"1"`
	Title   string               `json:"title" example:"データ指向アプリケーションデザイン"`
	Price   int64                `json:"price" example:"4600"`
	Status  int                  `json:"status" enums:"1,2" example:"2"`
	Authors []ListAuthorResponse `json:"authors"`
	ISBN    *string              `json:"isbn" example:"9784873118703"`
	// AmazonURL is derived from the ISBN; null when there is no ISBN (or no ISBN-10 form).
	AmazonURL *string `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	// CoverURL is the provider-hosted image; show it as-is with the credit for CoverSource.
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
} // @name BookListItemResponse

type ListResponse struct {
	Items  []ListItemResponse `json:"items"`
	Total  int                `json:"total" example:"1"`
	Limit  int                `json:"limit" example:"20"`
	Offset int                `json:"offset" example:"0"`
} // @name BookListResponse

// Handler converts HTTP ↔ UseCase only (no business logic).
type Handler struct {
	CreateUC bookcmd.CreateUsecase
	UpdateUC bookcmd.UpdateUsecase
	GetUC    bookqry.GetUsecase
	ListUC   bookqry.ListUsecase
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
}

// List godoc
// @Summary      書籍一覧を取得する
// @Description  書籍ID昇順。total は取得範囲外も含む総件数。amazonUrl は ISBN から導出し、coverUrl は提供元（coverSource）の画像で、楽天の場合は画面にクレジット表示が必要
// @Tags         books
// @Produce      json
// @Param        limit  query int false "取得件数（1〜100、既定20）"
// @Param        offset query int false "取得開始位置（0以上、既定0）"
// @Success      200 {object} ListResponse
// @Failure      400 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books [get]
func (h *Handler) List(c echo.Context) error {
	req, err := common.BindValidate[ListRequest](c)
	if err != nil {
		return err
	}
	limit, offset := domaincommon.DefaultListLimit, 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	out, err := h.ListUC.Execute(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toListResponse(out, limit, offset))
}

// Get godoc
// @Summary      書籍を取得する
// @Description  amazonUrl は ISBN から導出する。coverUrl は提供元（coverSource）がホストする画像で、楽天の場合は画面にクレジット表示が必要
// @Tags         books
// @Produce      json
// @Param        id path int true "書籍ID"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books/{id} [get]
func (h *Handler) Get(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.GetUC.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Create godoc
// @Summary      書籍を作成する
// @Description  書籍価格は0以上、著者は1人以上（重複不可・存在する著者のみ）
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        body body CreateRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books [post]
func (h *Handler) Create(c echo.Context) error {
	req, err := common.BindValidate[CreateRequest](c)
	if err != nil {
		return err
	}
	out, err := h.CreateUC.Execute(c.Request().Context(), bookcmd.CreateCommand{
		Title: req.Title, Price: *req.Price, AuthorIDs: req.AuthorIDs, Status: *req.Status, ISBN: req.ISBN,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Update godoc
// @Summary      書籍を更新する
// @Description  出版済みから未出版には変更できない。version が一致しない場合は 409
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id path int true "書籍ID"
// @Param        body body UpdateRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	req, err := common.BindValidate[UpdateRequest](c)
	if err != nil {
		return err
	}
	out, err := h.UpdateUC.Execute(c.Request().Context(), bookcmd.UpdateCommand{
		ID: id, Title: req.Title, Price: *req.Price, AuthorIDs: req.AuthorIDs, Status: *req.Status,
		ISBN: req.ISBN, Version: *req.Version,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

func toResponse(d *domainbook.BookDetail) Response {
	authors := make([]AuthorResponse, 0, len(d.Authors))
	for _, a := range d.Authors {
		authors = append(authors, AuthorResponse{ID: int64(a.ID), Name: a.Name, BirthDate: a.BirthDate, Version: a.Version})
	}
	return Response{
		ID: int64(d.ID), Title: d.Title, Price: d.Price, Authors: authors, Status: d.Status,
		ISBN: d.ISBN, AmazonURL: d.AmazonURL, CoverURL: d.CoverURL, CoverSource: d.CoverSource, Version: d.Version,
	}
}

func toListResponse(l *domainbook.BookList, limit, offset int) ListResponse {
	items := make([]ListItemResponse, 0, len(l.Items))
	for _, it := range l.Items {
		authors := make([]ListAuthorResponse, 0, len(it.Authors))
		for _, a := range it.Authors {
			authors = append(authors, ListAuthorResponse{ID: int64(a.ID), Name: a.Name})
		}
		items = append(items, ListItemResponse{
			ID: int64(it.ID), Title: it.Title, Price: it.Price, Status: it.Status, Authors: authors,
			ISBN: it.ISBN, AmazonURL: it.AmazonURL, CoverURL: it.CoverURL, CoverSource: it.CoverSource,
		})
	}
	return ListResponse{Items: items, Total: l.Total, Limit: limit, Offset: offset}
}
