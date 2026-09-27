package book

import (
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// uploadBodyLimit rejects oversized uploads (413) before they are read; it leaves
// room above the 5 MiB image limit for the multipart envelope.
const uploadBodyLimit = "6M"

type CreateRequest struct {
	Title     string  `json:"title" validate:"required,max=255" example:"人間失格"`
	Price     *int64  `json:"price" validate:"required,min=0,max=99999999" example:"1500"`
	AuthorIDs []int64 `json:"authorIds" validate:"required,min=1,dive,gt=0"`
	Status    *int    `json:"status" validate:"required,oneof=1 2" enums:"1,2" example:"1"`
	AmazonURL *string `json:"amazonUrl" validate:"omitempty,max=2048" example:"https://www.amazon.co.jp/dp/4101006059"`
} // @name CreateBookRequest

type UpdateRequest struct {
	Title     string  `json:"title" validate:"required,max=255" example:"人間失格"`
	Price     *int64  `json:"price" validate:"required,min=0,max=99999999" example:"1500"`
	AuthorIDs []int64 `json:"authorIds" validate:"required,min=1,dive,gt=0"`
	Status    *int    `json:"status" validate:"required,oneof=1 2" enums:"1,2" example:"1"`
	// Omitted or null clears the link (PUT replaces the whole book).
	AmazonURL *string `json:"amazonUrl" validate:"omitempty,max=2048" example:"https://www.amazon.co.jp/dp/4101006059"`
	Version   *int    `json:"version" validate:"required" example:"1"`
} // @name UpdateBookRequest

type AuthorResponse struct {
	ID        int64   `json:"id" example:"1"`
	Name      string  `json:"name" example:"太宰治"`
	BirthDate *string `json:"birthDate" example:"1909-06-19"`
	Version   int     `json:"version" example:"1"`
} // @name BookAuthorResponse

type Response struct {
	ID        int64            `json:"id" example:"1"`
	Title     string           `json:"title" example:"人間失格"`
	Price     int64            `json:"price" example:"1500"`
	Authors   []AuthorResponse `json:"authors"`
	Status    int              `json:"status" enums:"1,2" example:"1"`
	AmazonURL *string          `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4101006059"`
	// ImageURL is a presigned URL valid for 15 minutes; fetch the book again for a new one.
	ImageURL *string `json:"imageUrl" example:"http://localhost:4566/book-images/books/1/abc.png?X-Amz-Expires=900"`
	Version  int     `json:"version" example:"1"`
} // @name BookResponse

// Handler converts HTTP ↔ UseCase only (no business logic).
type Handler struct {
	CreateUC      bookcmd.CreateUsecase
	UpdateUC      bookcmd.UpdateUsecase
	UploadImageUC bookcmd.UploadImageUsecase
	GetUC         bookqry.GetUsecase
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.POST("/:id/image", h.UploadImage, middleware.BodyLimit(uploadBodyLimit))
}

// Get godoc
// @Summary      書籍を取得する
// @Description  imageUrl は15分間有効な署名付きURL
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

// UploadImage godoc
// @Summary      書籍の表紙画像をアップロードする
// @Description  JPEG / PNG / WebP、5MB以下。既存の画像は差し替える。種別はファイルの中身から判定する
// @Tags         books
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path     int  true "書籍ID"
// @Param        image formData file true "表紙画像"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      413 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books/{id}/image [post]
func (h *Handler) UploadImage(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	fh, err := c.FormFile("image")
	if err != nil {
		return &common.ValidationError{Fields: []common.FieldError{{Field: "image", Rule: "required"}}}
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	contentType, err := sniffContentType(f)
	if err != nil {
		return err
	}
	out, err := h.UploadImageUC.Execute(c.Request().Context(), bookcmd.UploadImageCommand{
		BookID: id, ContentType: contentType, Size: fh.Size, Body: f,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// sniffContentType decides the type from the bytes, not the client's Content-Type
// header, which is trivially spoofed; it rewinds f for the upload afterwards.
func sniffContentType(f io.ReadSeeker) (string, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(head[:n]), nil
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
		Title: req.Title, Price: *req.Price, AuthorIDs: req.AuthorIDs, Status: *req.Status, AmazonURL: req.AmazonURL,
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
		AmazonURL: req.AmazonURL, Version: *req.Version,
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
		AmazonURL: d.AmazonURL, ImageURL: d.ImageURL, Version: d.Version,
	}
}
