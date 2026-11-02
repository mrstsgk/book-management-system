package public

import (
	"net/http"

	"github.com/labstack/echo/v4"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

type ListRequest struct {
	Limit  *int `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int `json:"offset" query:"offset" validate:"omitempty,min=0"`
	// Q は書名・著者の部分一致。前後の空白を除いて空なら検索しない。
	// 文字数の上限は HTTP 層で見ない（トリム前の長さで弾くと、前後に空白の付いた100文字が400になるため）。VO が判定する。
	Q     string `json:"q" query:"q"`
	TagID *int64 `json:"tagId" query:"tagId" validate:"omitempty,min=1"`
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
	// CoverURL は提供元がホストする画像。
	CoverURL    *string `json:"coverUrl" example:"https://cover.openbd.jp/9784873118703.jpg"`
	CoverSource *string `json:"coverSource" enums:"openbd" example:"openbd"`
	// TitleOverride は自分で上書きした書名。上書きしていなければ null。
	TitleOverride *string  `json:"titleOverride" example:"徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版"`
	Summary       string   `json:"summary" example:"分散データの設計を体系的に学べる"`
	Tags          []string `json:"tags" example:"データベース,分散システム"`
	Comment       string   `json:"comment" example:"分散システムの設計を体系的に学べた"`
	Rating        int      `json:"rating" example:"5"`
	Version       int      `json:"version" example:"1"`
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
	CoverSource *string  `json:"coverSource" enums:"openbd" example:"openbd"`
	Rating      int      `json:"rating" example:"5"`
} // @name BookListItemResponse

type ListResponse struct {
	Items  []ListItemResponse `json:"items"`
	Total  int                `json:"total" example:"1"`
	Limit  int                `json:"limit" example:"20"`
	Offset int                `json:"offset" example:"0"`
} // @name BookListResponse

// Handler は、読んだ本の閲覧の API を処理する（誰でも使える）。
type Handler struct {
	ReadBookUsecase    bookqry.ReadBookUsecase
	SearchBooksUsecase bookqry.SearchBooksUsecase
}

// Register は、読んだ本の閲覧の API のルートを g に登録する。
func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.Search)
	g.GET("/:id", h.Read)
}

// Search は、読んだ本を検索する（GET /api/books）。
// @Summary      読んだ本の一覧を取得する
// @Description  新しく登録した順。q と tagId は同時に使える（両方を満たす本）。total は条件に合う、取得範囲外も含む総件数。感想の本文は詳細で返す。存在しない tagId は0件
// @Tags         books
// @Produce      json
// @Param        limit  query int    false "取得件数（1〜100、既定20）"
// @Param        offset query int    false "取得開始位置（0以上、既定0）"
// @Param        q      query string false "書名（上書きがあれば上書き）・著者の部分一致。大文字小文字を区別しない（100文字まで）"
// @Param        tagId  query int    false "この分野タグが付いた本だけにする（1以上）"
// @Success      200 {object} ListResponse
// @Failure      400 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/books [get]
func (h *Handler) Search(c echo.Context) error {
	req, err := httpserver.BindValidate[ListRequest](c)
	if err != nil {
		return err
	}
	limit, offset := bookdomain.DefaultListLimit, 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	out, err := h.SearchBooksUsecase.Execute(c.Request().Context(), bookqry.SearchBooksInput{Keyword: req.Q, TagID: req.TagID, Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	items := make([]ListItemResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, ListItemResponse{
			ID: int64(it.ID), ISBN: it.ISBN, Title: it.Title, Summary: it.Summary, Tags: it.Tags, Authors: it.Authors,
			AmazonURL: it.AmazonURL, CoverURL: it.CoverURL, CoverSource: it.CoverSource, Rating: it.Rating,
		})
	}
	return c.JSON(http.StatusOK, ListResponse{Items: items, Total: out.Total, Limit: limit, Offset: offset})
}

// Read は、読んだ本を 1 件参照する（GET /api/books/{id}）。
// @Summary      読んだ本を取得する
// @Tags         books
// @Produce      json
// @Param        id path int true "読んだ本のID"
// @Success      200 {object} Response
// @Failure      400 {object} httpserver.ErrorResponse
// @Failure      404 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/books/{id} [get]
func (h *Handler) Read(c echo.Context) error {
	id, err := httpserver.ParseID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.ReadBookUsecase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ToResponse(out))
}

// ToResponse は、読んだ本の詳細をレスポンスの形にする（管理側の登録・更新の応答も同じ形）。
func ToResponse(d *bookdomain.BookDetail) Response {
	return Response{
		ID: int64(d.ID), ISBN: d.ISBN, Title: d.Title, Authors: d.Authors, Publisher: d.Publisher,
		PublishedOn: d.PublishedOn, AmazonURL: d.AmazonURL, CoverURL: d.CoverURL, CoverSource: d.CoverSource,
		TitleOverride: d.TitleOverride, Summary: d.Summary, Tags: d.Tags, Comment: d.Comment, Rating: d.Rating, Version: d.Version,
	}
}
