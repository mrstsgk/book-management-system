package common

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const SessionCookieName = "admin_session"

// sessionCookieMaxAge は絶対期限（24 時間）と同じ。サーバー側が先に失効すれば Cookie が残っていても 401 になる。
const sessionCookieMaxAge = 86400

// SetSessionCookie はセッション ID を httpOnly Cookie で返す。Path=/api で画面のパスには送らせない。
// SameSite=Lax で他サイトからの POST には付かない（CSRF の一段目。二段目は RequireSameOrigin）。
// ponytail: Secure は付けない（ローカルの http でしか動かさない）。https で公開するなら Secure: true にする
func SetSessionCookie(c echo.Context, id string) {
	c.SetCookie(&http.Cookie{
		Name: SessionCookieName, Value: id, Path: "/api", MaxAge: sessionCookieMaxAge,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie はブラウザ側の Cookie を消す（サーバー側の行は LogoutUsecase が消す）。
func ClearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name: SessionCookieName, Value: "", Path: "/api", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}
