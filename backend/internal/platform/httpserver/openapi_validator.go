package httpserver

import (
	"cmp"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
)

// ValidationError は HTTP 境界の入力検証の違反（400。errors に項目ごとの違反を載せる）。
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "invalid json"
	}
	return "invalid request"
}

// ValidateRequest は、リクエストを OpenAPI の spec で検証し、違反があれば ValidationError（400）にする。
// spec を cmd から受け取るのは、platform が生成物（api/gen）を import しないため。
// spec に無いルート（/health など）はそのまま通す。
func ValidateRequest(spec *openapi3.T) (echo.MiddlewareFunc, error) {
	// servers は照合しない（同一オリジンで配信し、開発時は Vite が中継するため、Host が spec と一致しない）
	spec.Servers = nil
	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		return nil, err
	}
	opts := &openapi3filter.Options{
		// 最初の 1 件で止めず、違反をすべて返す（docs/architecture.md §4「バリデーションの全件返却」）
		MultiError: true,
		// 認証は middleware（RequireAdminSession）が先に済ませる。spec の security は文書のためだけに書く
		AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			route, pathParams, err := router.FindRoute(req)
			if err != nil {
				// spec に無いルートは Echo に任せる（404 / 405 / spec 外のハンドラ）
				return next(c)
			}
			err = openapi3filter.ValidateRequest(req.Context(), &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
				Options:    opts,
			})
			if err != nil {
				return toValidationError(err)
			}
			return next(c)
		}
	}, nil
}

// toValidationError は kin-openapi の検証エラーを ValidationError（field + rule）に変換する。
// JSON として読めない・必須なのに無いボディは項目なし（"invalid json"）、制約の違反は項目ごとに 1 件、
// どの項目か特定できない違反は errors の無い 400 にする。
func toValidationError(err error) error {
	var fields []FieldError
	malformed := false
	for _, e := range flatten(err) {
		re, ok := errors.AsType[*openapi3filter.RequestError](e)
		if !ok {
			continue
		}
		// 読めない・無いボディは、項目として返せるものが無い
		if _, isParse := errors.AsType[*openapi3filter.ParseError](re.Err); re.Parameter == nil && (isParse || errors.Is(re.Err, openapi3filter.ErrInvalidRequired)) {
			malformed = true
			continue
		}
		fields = append(fields, requestErrorFields(re)...)
	}
	if malformed {
		return &ValidationError{}
	}
	if len(fields) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	// kin-openapi の順序は実装の都合なので、テストとクライアントのために並びを決めておく（契約としては順不同）
	slices.SortFunc(fields, func(a, b FieldError) int {
		return cmp.Or(cmp.Compare(a.Field, b.Field), cmp.Compare(a.Rule, b.Rule))
	})
	return &ValidationError{Fields: fields}
}

func requestErrorFields(re *openapi3filter.RequestError) []FieldError {
	if re.Parameter != nil {
		rule := "type"
		if errors.Is(re.Err, openapi3filter.ErrInvalidRequired) {
			rule = "required"
		}
		for _, e := range flatten(re.Err) {
			if se, ok := errors.AsType[*openapi3.SchemaError](e); ok {
				rule = se.SchemaField
				break
			}
		}
		return []FieldError{{Field: re.Parameter.Name, Rule: rule}}
	}
	var fields []FieldError
	for _, e := range flatten(re.Err) {
		// 空のポインタはボディ全体（オブジェクトの代わりに配列など）。項目ではないので errors の無い 400 にする
		if se, ok := errors.AsType[*openapi3.SchemaError](e); ok && len(se.JSONPointer()) > 0 {
			fields = append(fields, FieldError{Field: strings.Join(se.JSONPointer(), "."), Rule: se.SchemaField})
		}
	}
	return fields
}

// flatten は入れ子の openapi3.MultiError を平らにする。
// errors.As ではなく型アサーションにするのは、MultiError に Unwrap される RequestError をそのまま残すため。
func flatten(err error) []error {
	me, ok := err.(openapi3.MultiError) //nolint:errorlint // 上の理由
	if !ok {
		return []error{err}
	}
	var out []error
	for _, e := range me {
		out = append(out, flatten(e)...)
	}
	return out
}
