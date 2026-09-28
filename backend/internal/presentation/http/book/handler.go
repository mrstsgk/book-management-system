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

type RegisterRequest struct {
	// ISBN は13桁または10桁（ハイフン可）。書誌と書影はこれで外部カタログから取得する。
	ISBN    string `json:"isbn" validate:"required,max=17" example:"9784873118703"`
	Summary string `json:"summary" validate:"required,max=100" example:"分散データの設計を体系的に学べる"`
	// TitleOverride は外部カタログの書名が実際と違うときに自分で付ける書名。省略・空（空白だけも）なら上書きしない。
	TitleOverride string `json:"titleOverride" validate:"max=255" example:""`
	// TagIDs は分野タグのID。0〜10個、実在するIDのみ指定できる。
	TagIDs  []int64 `json:"tagIds" validate:"max=10" example:"1,2"`
	Comment string  `json:"comment" validate:"required,max=5000" example:"分散システムの設計を体系的に学べた"`
	Rating  *int    `json:"rating" validate:"required,min=1,max=5" example:"5"`
} // @name RegisterBookRequest

type UpdateRequest struct {
	Summary string `json:"summary" validate:"required,max=100" example:"読み返して理解が深まった"`
	// TitleOverride は全体の置き換えなので、省略・空（空白だけも）なら上書きを外す。
	TitleOverride string `json:"titleOverride" validate:"max=255" example:""`
	// TagIDs は全体の置き換えなので、省略・空なら分野タグをすべて外す。
	TagIDs  []int64 `json:"tagIds" validate:"max=10" example:"1,2"`
	Comment string  `json:"comment" validate:"required,max=5000" example:"読み返して理解が深まった"`
	Rating  *int    `json:"rating" validate:"required,min=1,max=5" example:"5"`
	Version *int    `json:"version" validate:"required" example:"1"`
} // @name UpdateBookRequest

type ListRequest struct {
	Limit  *int `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

type Response struct {
	ID   int64  `json:"id" example:"1"`
	ISBN string `json:"isbn" example:"9784873118703"`
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title       string `json:"title" example:"データ指向アプリケーションデザイン"`
	Authors     string `json:"authors" example:"Kleppmann,Martin 斉藤,太郎 玉川,竜司"`
	Publisher   string `json:"publisher" example:"オーム社"`
	PublishedOn string `json:"publishedOn" example:"201907"`
	// AmazonURL は ISBN から導出する。ISBN-10 の形式が無ければ null。
	AmazonURL *string `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	// CoverURL は提供元がホストする画像。coverSource が rakuten なら画面にクレジット表示が必要。
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
	// CoverProductURL は楽天の商品ページ（楽天の書影のときだけ）。楽天の書影は画面でこのリンクと一緒に見せる必要がある。
	CoverProductURL *string `json:"coverProductUrl" example:"https://books.rakuten.co.jp/rb/15949390/"`
	// TitleOverride は自分で上書きした書名。上書きしていなければ null。
	TitleOverride *string  `json:"titleOverride" example:"徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版"`
	Summary       string   `json:"summary" example:"分散データの設計を体系的に学べる"`
	Tags          []string `json:"tags" example:"データベース,分散システム"`
	Comment       string   `json:"comment" example:"分散システムの設計を体系的に学べた"`
	Rating        int      `json:"rating" example:"5"`
	// RakutenDisabled は楽天から削除の指示を受けて楽天由来の情報を消した本（以後、楽天の書影は付かない）。
	RakutenDisabled bool `json:"rakutenDisabled" example:"false"`
	Version         int  `json:"version" example:"1"`
} // @name BookResponse

type ListItemResponse struct {
	ID   int64  `json:"id" example:"1"`
	ISBN string `json:"isbn" example:"9784873118703"`
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title       string   `json:"title" example:"データ指向アプリケーションデザイン"`
	Summary     string   `json:"summary" example:"分散データの設計を体系的に学べる"`
	Tags        []string `json:"tags" example:"データベース,分散システム"`
	Authors     string   `json:"authors" example:"Kleppmann,Martin 斉藤,太郎 玉川,竜司"`
	AmazonURL   *string  `json:"amazonUrl" example:"https://www.amazon.co.jp/dp/4873118700"`
	CoverURL    *string  `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string  `json:"coverSource" enums:"openbd,rakuten" example:"openbd"`
	// CoverProductURL は楽天の商品ページ（楽天の書影のときだけ）。
	CoverProductURL *string `json:"coverProductUrl" example:"https://books.rakuten.co.jp/rb/15949390/"`
	Rating          int     `json:"rating" example:"5"`
} // @name BookListItemResponse

type ListResponse struct {
	Items  []ListItemResponse `json:"items"`
	Total  int                `json:"total" example:"1"`
	Limit  int                `json:"limit" example:"20"`
	Offset int                `json:"offset" example:"0"`
} // @name BookListResponse

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	RegisterUC bookcmd.RegisterUsecase
	UpdateUC   bookcmd.UpdateUsecase
	DeleteUC   bookcmd.DeleteUsecase
	// DisableRakutenUC は楽天から削除の指示を受けた本の楽天由来の情報を消す。
	DisableRakutenUC bookcmd.DisableRakutenUsecase
	GetUC            bookqry.GetUsecase
	ListUC           bookqry.ListUsecase
	// AdminOnly は登録・更新・削除に掛ける認証（閲覧は誰でもできる）。
	AdminOnly echo.MiddlewareFunc
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.RegisterBook, h.AdminOnly)
	g.PUT("/:id", h.Update, h.AdminOnly)
	g.DELETE("/:id", h.Delete, h.AdminOnly)
	g.DELETE("/:id/rakuten", h.DisableRakuten, h.AdminOnly)
}

// List godoc
// @Summary      読んだ本の一覧を取得する
// @Description  新しく登録した順。total は取得範囲外も含む総件数。感想の本文は詳細で返す
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
	items := make([]ListItemResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, ListItemResponse{
			ID: int64(it.ID), ISBN: it.ISBN, Title: it.Title, Summary: it.Summary, Tags: it.Tags, Authors: it.Authors,
			AmazonURL: it.AmazonURL, CoverURL: it.CoverURL, CoverSource: it.CoverSource, CoverProductURL: it.CoverProductURL, Rating: it.Rating,
		})
	}
	return c.JSON(http.StatusOK, ListResponse{Items: items, Total: out.Total, Limit: limit, Offset: offset})
}

// Get godoc
// @Summary      読んだ本を取得する
// @Tags         books
// @Produce      json
// @Param        id path int true "読んだ本のID"
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

// RegisterBook godoc
// @Summary      読んだ本を登録する（自分だけ）
// @Description  書誌は ISBN で openBD から取得し、書影が無ければ楽天ブックスで補う。openBD に無い ISBN は 400、同じ ISBN の登録済みは 409
// @Tags         books
// @Accept       json
// @Produce      json
// @Security     AdminToken
// @Param        body body RegisterRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books [post]
func (h *Handler) RegisterBook(c echo.Context) error {
	req, err := common.BindValidate[RegisterRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RegisterUC.Execute(c.Request().Context(), bookcmd.RegisterCommand{
		ISBN: req.ISBN, Summary: req.Summary, TitleOverride: req.TitleOverride, TagIDs: req.TagIDs, Comment: req.Comment, Rating: *req.Rating,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Update godoc
// @Summary      感想と評価を更新する（自分だけ）
// @Description  あわせて書誌と書影を外部カタログから取り直す（取り直せなければ今のまま）。titleOverride を省略・空にすると上書きを外す。version が一致しない場合は 409
// @Tags         books
// @Accept       json
// @Produce      json
// @Security     AdminToken
// @Param        id   path int           true "読んだ本のID"
// @Param        body body UpdateRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
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
		ID: id, Summary: req.Summary, TitleOverride: req.TitleOverride, TagIDs: req.TagIDs, Comment: req.Comment, Rating: *req.Rating, Version: *req.Version,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Delete godoc
// @Summary      読んだ本を削除する（自分だけ）
// @Tags         books
// @Security     AdminToken
// @Param        id path int true "読んだ本のID"
// @Success      204
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := h.DeleteUC.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// DisableRakuten godoc
// @Summary      楽天由来の情報を消す（自分だけ）
// @Description  楽天から削除の指示を受けた本の楽天の書影・商品ページを消し、以後この本には楽天の書影を付けない。楽天の情報を持たない本でも 204。version は受け取らない（削除の指示には最後に読んだ内容に関係なく従うため）
// @Tags         books
// @Security     AdminToken
// @Param        id path int true "読んだ本のID"
// @Success      204
// @Failure      400 {object} common.ErrorResponse
// @Failure      401 {object} common.ErrorResponse
// @Failure      404 {object} common.ErrorResponse
// @Failure      409 {object} common.ErrorResponse
// @Failure      500 {object} common.ErrorResponse
// @Router       /api/books/{id}/rakuten [delete]
func (h *Handler) DisableRakuten(c echo.Context) error {
	id, err := common.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := h.DisableRakutenUC.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func toResponse(d *domainbook.BookDetail) Response {
	return Response{
		ID: int64(d.ID), ISBN: d.ISBN, Title: d.Title, Authors: d.Authors, Publisher: d.Publisher,
		PublishedOn: d.PublishedOn, AmazonURL: d.AmazonURL, CoverURL: d.CoverURL, CoverSource: d.CoverSource, CoverProductURL: d.CoverProductURL,
		TitleOverride: d.TitleOverride, Summary: d.Summary, Tags: d.Tags, Comment: d.Comment, Rating: d.Rating, RakutenDisabled: d.RakutenDisabled, Version: d.Version,
	}
}
