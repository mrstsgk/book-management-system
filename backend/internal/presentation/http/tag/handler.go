package tag

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

type RegisterRequest struct {
	Name string `json:"name" validate:"required,max=30" example:"データベース"`
} // @name RegisterTagRequest

type RenameRequest struct {
	Name    string `json:"name" validate:"required,max=30" example:"分散システム"`
	Version *int   `json:"version" validate:"required" example:"1"`
} // @name RenameTagRequest

type Response struct {
	ID      int64  `json:"id" example:"1"`
	Name    string `json:"name" example:"データベース"`
	Version int    `json:"version" example:"1"`
} // @name TagResponse

type ListResponse struct {
	Items []Response `json:"items"`
} // @name TagListResponse

type BookCountResponse struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"データベース"`
	BookCount int    `json:"bookCount" example:"3"`
} // @name TagBookCountResponse

type BookCountListResponse struct {
	Items []BookCountResponse `json:"items"`
} // @name TagBookCountListResponse

// Handler は HTTP と UseCase の変換だけを行う（業務ロジックは持たない）。
type Handler struct {
	RegisterTagUsecase tagcmd.RegisterTagUsecase
	RenameTagUsecase   tagcmd.RenameTagUsecase
	DeleteTagUsecase   tagcmd.DeleteTagUsecase
	SearchTagsUsecase  tagqry.SearchTagsUsecase
	CountBooksUsecase  tagqry.CountBooksUsecase
	// AdminOnly は追加・改名・削除に掛ける認証（閲覧は誰でもできる）。
	AdminOnly echo.MiddlewareFunc
}

// Register は、分野タグの API のルートを g に登録する。
func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.Search)
	g.GET("/counts", h.CountBooks)
	g.POST("", h.RegisterTag, h.AdminOnly)
	g.PUT("/:id", h.Rename, h.AdminOnly)
	g.DELETE("/:id", h.Delete, h.AdminOnly)
}

// Search は、分野タグの一覧を参照する（GET /api/tags）。
// @Summary      分野タグの一覧を取得する
// @Tags         tags
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/tags [get]
func (h *Handler) Search(c echo.Context) error {
	out, err := h.SearchTagsUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]Response, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, Response{ID: int64(it.ID), Name: it.Name, Version: it.Version})
	}
	return c.JSON(http.StatusOK, ListResponse{Items: items})
}

// CountBooks は、分野タグごとの冊数を返す（GET /api/tags/counts）。
// @Summary      分野タグごとの冊数を取得する
// @Description  本が1冊も付いていないタグは含めない。冊数の多い順、同数ならタグ名順
// @Tags         tags
// @Produce      json
// @Success      200 {object} BookCountListResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/tags/counts [get]
func (h *Handler) CountBooks(c echo.Context) error {
	out, err := h.CountBooksUsecase.Execute(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]BookCountResponse, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, BookCountResponse{ID: int64(it.ID), Name: it.Name, BookCount: it.BookCount})
	}
	return c.JSON(http.StatusOK, BookCountListResponse{Items: items})
}

// RegisterTag は、分野タグを追加する（POST /api/tags）。
// @Summary      分野タグを追加する（自分だけ）
// @Description  同じ名前のタグは追加できない。入力・VO・相関ルールの違反は 400 の errors[{field, rule}] で返す（rule: validate のキーワード / invalid / tagNameDuplicated）
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} httpserver.ErrorResponse "入力・VO・相関ルールの違反（errors に全件）"
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      409 {object} httpserver.ErrorResponse "同じ名前のタグと同時に追加されたとき（一意制約。通常は 400 の tagNameDuplicated）"
// @Failure      500 {object} httpserver.ErrorResponse
// @x-correlation-rules [{"id":"tagNameDuplicated","field":"name","rule":"タグ名は他のタグと重複しない","message":"MSG-A04"}]
// @Router       /api/tags [post]
func (h *Handler) RegisterTag(c echo.Context) error {
	req, err := httpserver.BindValidate[RegisterRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RegisterTagUsecase.Execute(c.Request().Context(), tagcmd.RegisterTagCommand{Name: req.Name})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Rename は、分野タグの名前を変える（PUT /api/tags/{id}）。
// @Summary      分野タグの名前を変更する（自分だけ）
// @Description  自分以外に同じ名前のタグがあると変更できない。入力・VO・相関ルールの違反は 400 の errors[{field, rule}] で返す（rule: validate のキーワード / invalid / tagNameDuplicated）。version が一致しない場合は 409
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path int           true "タグのID"
// @Param        body body RenameRequest true "body"
// @Success      200 {object} Response
// @Failure      400 {object} httpserver.ErrorResponse "入力・VO・相関ルールの違反（errors に全件）"
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      404 {object} httpserver.ErrorResponse
// @Failure      409 {object} httpserver.ErrorResponse "version が一致しない（他で先に更新された）"
// @Failure      500 {object} httpserver.ErrorResponse
// @x-correlation-rules [{"id":"tagNameDuplicated","field":"name","rule":"タグ名は自分以外のタグと重複しない","message":"MSG-A04"}]
// @Router       /api/tags/{id} [put]
func (h *Handler) Rename(c echo.Context) error {
	id, err := httpserver.ParseID(c, "id")
	if err != nil {
		return err
	}
	req, err := httpserver.BindValidate[RenameRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RenameTagUsecase.Execute(c.Request().Context(), tagcmd.RenameTagCommand{ID: id, Name: req.Name, Version: *req.Version})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toResponse(out))
}

// Delete は、分野タグを削除する（DELETE /api/tags/{id}）。
// @Summary      分野タグを削除する（自分だけ）
// @Description  付いていた本からは自動で外れる
// @Tags         tags
// @Param        id path int true "タグのID"
// @Success      204
// @Failure      400 {object} httpserver.ErrorResponse
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      404 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/tags/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := httpserver.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := h.DeleteTagUsecase.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func toResponse(v *tagcmd.TagView) Response {
	return Response{ID: v.ID, Name: v.Name, Version: v.Version}
}
