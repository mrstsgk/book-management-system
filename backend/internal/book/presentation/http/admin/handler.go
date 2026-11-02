package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"

	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
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

// Handler は、読んだ本の登録・更新・削除の API を処理する（自分だけ）。
type Handler struct {
	RegisterBookUsecase bookcmd.RegisterBookUsecase
	UpdateBookUsecase   bookcmd.UpdateBookUsecase
	DeleteBookUsecase   bookcmd.DeleteBookUsecase
	// AdminOnly は書き込みに掛ける認証（セッションと Origin の検証）。
	AdminOnly echo.MiddlewareFunc
}

// Register は、読んだ本の書き込みの API のルートを g に登録する。
func (h *Handler) Register(g *echo.Group) {
	g.POST("", h.RegisterBook, h.AdminOnly)
	g.PUT("/:id", h.Update, h.AdminOnly)
	g.DELETE("/:id", h.Delete, h.AdminOnly)
}

// RegisterBook は、読んだ本を登録する（POST /api/books）。
// @Summary      読んだ本を登録する（自分だけ）
// @Description  書誌と書影は ISBN で openBD から取得する。入力・VO・相関ルールの違反は独立して判定できるものをすべて 400 の errors[{field, rule}] で返す（rule: validate のキーワード / isbn の format / isbnDuplicated / isbnNotInCatalog / tagIdsNotFound / tagIdsDuplicated）。外部カタログの照会は他の違反が無いときだけ行う
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "body"
// @Success      200 {object} internal_book_presentation_http_public.Response
// @Failure      400 {object} httpserver.ErrorResponse "入力・VO・相関ルールの違反（errors に全件）"
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      409 {object} httpserver.ErrorResponse "同じ ISBN の本と同時に登録されたとき（一意制約。通常は 400 の isbnDuplicated）"
// @Failure      500 {object} httpserver.ErrorResponse
// @x-correlation-rules [{"id":"isbnDuplicated","field":"isbn","rule":"登録する ISBN の本は未登録である","message":"MSG-A01"},{"id":"isbnNotInCatalog","field":"isbn","rule":"登録する ISBN は外部カタログに存在する","message":"MSG-A02"},{"id":"tagIdsNotFound","field":"tagIds","rule":"指定した分野タグ ID はすべて存在する","message":"MSG-A03"},{"id":"tagIdsDuplicated","field":"tagIds","rule":"分野タグ ID に重複が無い","message":"（共通）"}]
// @Router       /api/books [post]
func (h *Handler) RegisterBook(c echo.Context) error {
	req, err := httpserver.BindValidate[RegisterRequest](c)
	if err != nil {
		return err
	}
	out, err := h.RegisterBookUsecase.Execute(c.Request().Context(), bookcmd.RegisterBookCommand{
		ISBN: req.ISBN, Summary: req.Summary, TitleOverride: req.TitleOverride, TagIDs: req.TagIDs, Comment: req.Comment, Rating: *req.Rating,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bookpublic.ToResponse(out))
}

// Update は、読んだ本の感想・評価・書名の上書き・分野タグを更新する（PUT /api/books/{id}）。
// @Summary      感想と評価を更新する（自分だけ）
// @Description  あわせて書誌と書影を外部カタログから取り直す（取り直せなければ今のまま）。titleOverride を省略・空にすると上書きを外す。入力・VO・相関ルールの違反は独立して判定できるものをすべて 400 の errors[{field, rule}] で返す（rule: validate のキーワード / tagIdsNotFound / tagIdsDuplicated）。version が一致しない場合は 409
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id   path int           true "読んだ本のID"
// @Param        body body UpdateRequest true "body"
// @Success      200 {object} internal_book_presentation_http_public.Response
// @Failure      400 {object} httpserver.ErrorResponse "入力・VO・相関ルールの違反（errors に全件）"
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      404 {object} httpserver.ErrorResponse
// @Failure      409 {object} httpserver.ErrorResponse "version が一致しない（他で先に更新された）"
// @Failure      500 {object} httpserver.ErrorResponse
// @x-correlation-rules [{"id":"tagIdsNotFound","field":"tagIds","rule":"指定した分野タグ ID はすべて存在する","message":"MSG-A03"},{"id":"tagIdsDuplicated","field":"tagIds","rule":"分野タグ ID に重複が無い","message":"（共通）"}]
// @Router       /api/books/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	id, err := httpserver.ParseID(c, "id")
	if err != nil {
		return err
	}
	req, err := httpserver.BindValidate[UpdateRequest](c)
	if err != nil {
		return err
	}
	out, err := h.UpdateBookUsecase.Execute(c.Request().Context(), bookcmd.UpdateBookCommand{
		ID: id, Summary: req.Summary, TitleOverride: req.TitleOverride, TagIDs: req.TagIDs, Comment: req.Comment, Rating: *req.Rating, Version: *req.Version,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bookpublic.ToResponse(out))
}

// Delete は、読んだ本を削除する（DELETE /api/books/{id}）。
// @Summary      読んだ本を削除する（自分だけ）
// @Tags         books
// @Param        id path int true "読んだ本のID"
// @Success      204
// @Failure      400 {object} httpserver.ErrorResponse
// @Failure      401 {object} httpserver.ErrorResponse
// @Failure      403 {object} httpserver.ErrorResponse
// @Failure      404 {object} httpserver.ErrorResponse
// @Failure      500 {object} httpserver.ErrorResponse
// @Router       /api/books/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := httpserver.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := h.DeleteBookUsecase.Execute(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
