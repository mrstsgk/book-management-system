package common

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RequireSameOrigin は状態を変える要求（GET / HEAD 以外）の Origin が「自分のスキーム://Host」と一致しなければ 403 にする。
// SameSite=Lax の Cookie だけでも他サイトからの POST には付かないが、将来の既定値の変更に備えて二段目として置く。
// ブラウザは POST に必ず Origin を付けるので、無ければ（または "null" なら）拒否してよい。Referer は見ない。
func RequireSameOrigin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Method == http.MethodGet || req.Method == http.MethodHead {
				return next(c)
			}
			if req.Header.Get("Origin") != c.Scheme()+"://"+req.Host {
				return echo.NewHTTPError(http.StatusForbidden, "forbidden")
			}
			return next(c)
		}
	}
}
