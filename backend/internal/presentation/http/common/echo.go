package common

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = HTTPErrorHandler
	e.Use(middleware.RequestID())
	e.Use(requestLogger())
	// パニックでもリクエスト ID 付きのアクセスログが残るよう、ロガーの内側に置く。
	e.Use(middleware.Recover())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return e
}

// requestLogger はリクエストごとに 1 行のアクセスログを slog に書く。
// エラーの詳細は HTTPErrorHandler が warn/error で出すので、ここは info のまま。
// クエリ文字列はトークンや個人的な検索語を含みうるため記録しない。
func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:   true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURIPath:   true,
		LogStatus:    true,
		LogRequestID: true,
		// 記録するステータスが最終のものになるよう、先にエラーハンドラを走らせる。
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			slog.Info("request",
				"request_id", v.RequestID,
				"method", v.Method,
				"path", v.URIPath,
				"status", v.Status,
				"latency", v.Latency.String(),
				"remote_ip", v.RemoteIP,
			)
			return nil
		},
	})
}
