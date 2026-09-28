package common_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

type bindTarget struct {
	Title string `json:"title" validate:"required,max=50"`
}

func newBindContext(body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestBindValidate_ValidBody(t *testing.T) {
	t.Parallel()
	c, _ := newBindContext(`{"title":"人間失格"}`)

	got, err := common.BindValidate[bindTarget](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "人間失格" {
		t.Fatalf("Title = %q, want 人間失格", got.Title)
	}
}

func TestBindValidate_MissingRequiredField(t *testing.T) {
	t.Parallel()
	c, _ := newBindContext(`{}`)

	_, err := common.BindValidate[bindTarget](c)
	var ve *common.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %T, want *ValidationError", err)
	}
	want := []common.FieldError{{Field: "title", Rule: "required"}}
	if len(ve.Fields) != 1 || ve.Fields[0] != want[0] {
		t.Fatalf("Fields = %+v, want %+v", ve.Fields, want)
	}
}

func TestBindValidate_MalformedJSON(t *testing.T) {
	t.Parallel()
	c, _ := newBindContext(`{"title":`)

	_, err := common.BindValidate[bindTarget](c)
	var ve *common.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %T, want *ValidationError", err)
	}
	if len(ve.Fields) != 0 {
		t.Fatalf("Fields = %+v, want empty (bind failure, not a field validation failure)", ve.Fields)
	}
}

type queryTarget struct {
	Limit *int `json:"limit" query:"limit" validate:"omitempty,min=1"`
}

func TestBindValidate_QueryParams(t *testing.T) {
	t.Parallel()
	newQueryContext := func(query string) echo.Context {
		req := httptest.NewRequest(http.MethodGet, "/?"+query, nil)
		return echo.New().NewContext(req, httptest.NewRecorder())
	}

	t.Run("数値のクエリはバインドされる", func(t *testing.T) {
		t.Parallel()
		got, err := common.BindValidate[queryTarget](newQueryContext("limit=5"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Limit == nil || *got.Limit != 5 {
			t.Fatalf("Limit = %v, want 5", got.Limit)
		}
	})

	t.Run("型が合わないクエリはフィールドなしのValidationError", func(t *testing.T) {
		t.Parallel()
		_, err := common.BindValidate[queryTarget](newQueryContext("limit=abc"))
		var ve *common.ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %T, want *ValidationError", err)
		}
		if len(ve.Fields) != 0 || ve.Error() != "malformed request" {
			t.Fatalf("got fields=%+v message=%q, want no fields and \"malformed request\"", ve.Fields, ve.Error())
		}
	})

	t.Run("バリデーション違反はJSON名のフィールドエラー", func(t *testing.T) {
		t.Parallel()
		_, err := common.BindValidate[queryTarget](newQueryContext("limit=0"))
		var ve *common.ValidationError
		if !errors.As(err, &ve) || len(ve.Fields) != 1 || ve.Fields[0] != (common.FieldError{Field: "limit", Rule: "min"}) {
			t.Fatalf("got %v, want limit/min field error", err)
		}
	})
}
