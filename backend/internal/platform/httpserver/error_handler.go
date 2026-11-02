package httpserver

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// ErrorResponse は、エラーのときに返すボディ（openapi/common/schemas/ErrorResponse.yaml と同じ形）。
// platform はチャネルの生成物を import しないので手で書く。形のずれは、レスポンスを spec で検証するテストで見つける。
type ErrorResponse struct {
	// Message は開発者・ログ向けの説明。画面には出さない。
	Message string `json:"message"`
	// Code は業務ルールの違反のコード（x-error-codes）。5xx には付かない。
	Code string `json:"code,omitempty"`
	// Params は文言に差し込む値（BE しか知らない値だけ）。
	Params map[string]any `json:"params,omitempty"`
	// Errors は項目ごとの違反。入力検証・VO・相関ルールの違反（400）だけに付く。
	Errors []FieldError `json:"errors,omitempty"`
}

// FieldError は、1 つの項目の 1 つの違反。
type FieldError struct {
	// Field は違反した項目の JSON 名（入れ子はドット区切り、パラメータはパラメータ名）。
	Field string `json:"field"`
	// Rule は OpenAPI の制約キーワード、VO の違反（format / invalid）、相関ルールの id のいずれか。
	Rule string `json:"rule"`
}

// 5xx では err.Error() を返さない（ドライバやネットワークのエラーはホスト名や資格情報を含む）。
const internalMessage = "internal"

// HTTPErrorHandler はエラーを HTTP レスポンスに変換する唯一の場所。Handler は return err するだけでよい。
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status, res := toErrorResponse(err)

	req := c.Request()
	attrs := []any{
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
		"error", err, "status", status, "method", req.Method, "path", req.URL.Path,
	}
	if status >= http.StatusInternalServerError {
		slog.Error("request failed", attrs...)
	} else {
		slog.Warn("request rejected", attrs...)
	}

	if req.Method == http.MethodHead {
		err = c.NoContent(status)
	} else {
		err = c.JSON(status, res)
	}
	if err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

// sentinelStatuses は domain の sentinel error と HTTP ステータスの対応。
var sentinelStatuses = []struct {
	err    error
	status int
}{
	{domainerr.ErrNotFound, http.StatusNotFound},
	{domainerr.ErrInvalid, http.StatusBadRequest},
	{domainerr.ErrConflict, http.StatusConflict},
	{domainerr.ErrUnauthorized, http.StatusUnauthorized},
	{domainerr.ErrTooManyAttempts, http.StatusTooManyRequests},
}

// toErrorResponse はエラーを HTTP ステータスとボディへ変換する: 入力検証の違反と UseCase が集めた違反は
// 項目ごとの errors 付きの 400、既知の分類はそれぞれのステータス（業務ルールの違反ならコードと差し込む値も）、
// echo の HTTPError はそのコード（5xx は説明を隠す）、それ以外は汎用の 500。
func toErrorResponse(err error) (int, ErrorResponse) {
	if ve, ok := errors.AsType[*ValidationError](err); ok {
		return http.StatusBadRequest, ErrorResponse{Message: ve.Error(), Errors: ve.Fields}
	}
	// UseCase が集めた相関ルール・VO の違反も、入力検証と同じ形（400 + errors[{field, rule}]）で返す
	if vs, ok := errors.AsType[domainerr.Violations](err); ok {
		fields := make([]FieldError, 0, len(vs))
		for _, v := range vs {
			fields = append(fields, FieldError{Field: v.Field, Rule: v.Rule})
		}
		return http.StatusBadRequest, ErrorResponse{Message: "invalid request", Errors: fields}
	}
	if status, ok := sentinelStatus(err); ok {
		res := ErrorResponse{Message: err.Error()}
		// 業務ルールの違反のコードと差し込む値は、包まれていても取り出す
		if ce, ok := errors.AsType[*domainerr.CodedError](err); ok {
			res.Code, res.Params = ce.Code, ce.Params
		}
		return status, res
	}
	if he, ok := errors.AsType[*echo.HTTPError](err); ok {
		if he.Code >= http.StatusInternalServerError {
			return he.Code, ErrorResponse{Message: internalMessage}
		}
		msg, ok := he.Message.(string)
		if !ok || msg == "" {
			msg = http.StatusText(he.Code)
		}
		return he.Code, ErrorResponse{Message: msg}
	}
	return http.StatusInternalServerError, ErrorResponse{Message: internalMessage}
}

func sentinelStatus(err error) (int, bool) {
	for _, m := range sentinelStatuses {
		if errors.Is(err, m.err) {
			return m.status, true
		}
	}
	return 0, false
}
