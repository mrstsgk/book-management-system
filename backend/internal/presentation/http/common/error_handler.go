package common

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ErrorResponse struct {
	Message string `json:"message" example:"invalid"`
	// Field-level details; only set for request validation errors.
	Errors []FieldError `json:"errors,omitempty"`
} // @name ErrorResponse

type FieldError struct {
	Field string `json:"field" example:"title"`
	Rule  string `json:"rule" example:"required"`
} // @name FieldError

// 5xx never echoes err.Error(): driver / network errors carry hosts and credentials.
const internalMessage = "internal"

// HTTPErrorHandler is the single place that maps errors to HTTP responses.
// Handlers just `return err`.
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
	{domaincommon.ErrNotFound, http.StatusNotFound},
	{domaincommon.ErrInvalid, http.StatusBadRequest},
	{domaincommon.ErrConflict, http.StatusConflict},
	{domaincommon.ErrUnauthorized, http.StatusUnauthorized},
	{domaincommon.ErrTooManyAttempts, http.StatusTooManyRequests},
}

// toErrorResponse はエラーを HTTP ステータスとボディへ変換する: validation
// エラーはフィールド詳細付き 400、既知の domain sentinel はそれぞれのステータス、
// echo の HTTPError はそのコード（5xx はメッセージを隠す）、それ以外は汎用 500。
func toErrorResponse(err error) (int, ErrorResponse) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return http.StatusBadRequest, ErrorResponse{Message: ve.Error(), Errors: ve.Fields}
	}
	for _, m := range sentinelStatuses {
		if errors.Is(err, m.err) {
			return m.status, ErrorResponse{Message: err.Error()}
		}
	}
	var he *echo.HTTPError
	if errors.As(err, &he) {
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
