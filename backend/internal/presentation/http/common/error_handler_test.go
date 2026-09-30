package common_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestHTTPErrorHandler(t *testing.T) {
	silenceLog(t)

	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
		wantErrors  []common.FieldError
		// notInMessage must not leak into the response body.
		notInMessage string
	}{
		{
			name:        "ErrNotFound is 404",
			err:         domaincommon.ErrNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "not found",
		},
		{
			name:        "wrapped ErrInvalid is 400 with its message",
			err:         fmt.Errorf("%w: title must be 1..50 characters", domaincommon.ErrInvalid),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid: title must be 1..50 characters",
		},
		{
			name:        "ErrConflict is 409",
			err:         domaincommon.ErrConflict,
			wantStatus:  http.StatusConflict,
			wantMessage: "conflict",
		},
		{
			name:        "ErrUnauthorized is 401",
			err:         domaincommon.ErrUnauthorized,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "unauthorized",
		},
		{
			name:        "ErrTooManyAttempts is 429",
			err:         domaincommon.ErrTooManyAttempts,
			wantStatus:  http.StatusTooManyRequests,
			wantMessage: "too many attempts",
		},
		{
			name:        "ValidationError is 400 with field errors",
			err:         &common.ValidationError{Fields: []common.FieldError{{Field: "title", Rule: "required"}}},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid request",
			wantErrors:  []common.FieldError{{Field: "title", Rule: "required"}},
		},
		{
			name:        "echo 4xx keeps its code and message",
			err:         echo.NewHTTPError(http.StatusBadRequest, "invalid id"),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid id",
		},
		{
			name:        "echo 404 without message uses status text",
			err:         echo.ErrNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "Not Found",
		},
		{
			name:         "echo 5xx hides its message",
			err:          echo.NewHTTPError(http.StatusServiceUnavailable, "pg: connection refused"),
			wantStatus:   http.StatusServiceUnavailable,
			wantMessage:  "internal",
			notInMessage: "connection refused",
		},
		{
			name:         "unknown error is 500 without internals",
			err:          errors.New("dial tcp 10.0.0.1:5432: connect: connection refused"),
			wantStatus:   http.StatusInternalServerError,
			wantMessage:  "internal",
			notInMessage: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveError(t, http.MethodGet, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.notInMessage != "" && strings.Contains(rec.Body.String(), tt.notInMessage) {
				t.Errorf("body leaks %q: %s", tt.notInMessage, rec.Body.String())
			}
			var got common.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			if !reflect.DeepEqual(got.Errors, tt.wantErrors) {
				t.Errorf("errors = %+v, want %+v", got.Errors, tt.wantErrors)
			}
		})
	}

	t.Run("HEAD has no body", func(t *testing.T) {
		rec := serveError(t, http.MethodHead, domaincommon.ErrNotFound)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("already committed response is left untouched", func(t *testing.T) {
		silenceLog(t)
		e := common.NewEcho()
		e.GET("/boom", func(c echo.Context) error {
			if err := c.String(http.StatusOK, "partial"); err != nil {
				return err
			}
			return domaincommon.ErrNotFound
		})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d (handler's own commit must win)", rec.Code, http.StatusOK)
		}
		if rec.Body.String() != "partial" {
			t.Errorf("body = %q, want %q", rec.Body.String(), "partial")
		}
	})
}

// serveError routes a request through NewEcho so the registered error handler is exercised.
func serveError(t *testing.T, method string, err error) *httptest.ResponseRecorder {
	t.Helper()
	e := common.NewEcho()
	e.Add(method, "/boom", func(echo.Context) error { return err })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, "/boom", nil))
	return rec
}
