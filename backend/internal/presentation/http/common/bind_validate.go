package common

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Report JSON names so clients can match errors to the fields they sent.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
	return v
}()

// ValidationError is a request bind / validate failure (→ 400 with Fields).
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		// Bind covers both the JSON body and query parameters, so don't name either one.
		return "malformed request"
	}
	return "invalid request"
}

// BindValidate はリクエストボディを T にバインドし構造体バリデーションを行う。
// バインド失敗や各 validator.ValidationErrors エントリを ValidationError
// （field + rule）に変換し、HTTPErrorHandler がそれを 400 に変換する。
func BindValidate[T any](c echo.Context) (T, error) {
	var req T
	if err := c.Bind(&req); err != nil {
		return req, &ValidationError{}
	}
	if err := validate.Struct(&req); err != nil {
		var ves validator.ValidationErrors
		if !errors.As(err, &ves) {
			return req, err
		}
		fields := make([]FieldError, 0, len(ves))
		for _, fe := range ves {
			fields = append(fields, FieldError{Field: fe.Field(), Rule: fe.Tag()})
		}
		return req, &ValidationError{Fields: fields}
	}
	return req, nil
}
