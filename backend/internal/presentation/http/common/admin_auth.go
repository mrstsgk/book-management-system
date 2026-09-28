package common

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireAdminToken は書き込み系の API を自分だけが使えるようにする。Authorization: Bearer <token> が
// token と一致しなければ 401 にする。token が空なら常に拒否する（設定漏れで誰でも書き込めるようにしないため）。
func RequireAdminToken(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			got, ok := strings.CutPrefix(c.Request().Header.Get(echo.HeaderAuthorization), "Bearer ")
			// 比較にかかる時間からトークンを推測されないよう、定数時間で比較する
			if token == "" || !ok || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
