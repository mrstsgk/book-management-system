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
