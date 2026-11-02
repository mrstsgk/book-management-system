package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const SessionCookieName = "admin_session"

// cookiePath は管理の API（/api/admin 配下）だけに Cookie を送らせるためのパス。
const cookiePath = "/api/admin"

// sessionCookieMaxAge は絶対期限（24 時間）と同じ。サーバー側が先に失効すれば Cookie が残っていても 401 になる。
const sessionCookieMaxAge = 86400

// SetSessionCookie はセッション ID を httpOnly Cookie で返す。Path=/api/admin で、閲覧の API と画面のパスには送らせない。
// SameSite=Lax で他サイトからの POST には付かない（CSRF の一段目。二段目は RequireSameOrigin）。
// ponytail: Secure は付けない（ローカルの http でしか動かさない）。https で公開するなら Secure: true にする
func SetSessionCookie(c echo.Context, id string) {
	c.SetCookie(&http.Cookie{ //nolint:gosec // Secure を付けない理由は上のコメント（ローカルの http 限定）

		Name: SessionCookieName, Value: id, Path: cookiePath, MaxAge: sessionCookieMaxAge,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie はブラウザ側の Cookie を消す（サーバー側の行は LogoutUsecase が消す）。
func ClearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{ //nolint:gosec // 消すだけの Cookie。Secure を付けない理由は SetSessionCookie と同じ

		Name: SessionCookieName, Value: "", Path: cookiePath, MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}
