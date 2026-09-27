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
	// Inside the logger so panics still get an access log line with the request ID.
	e.Use(middleware.Recover())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return e
}

// requestLogger writes one access log line per request via slog.
// Error details are logged by HTTPErrorHandler at warn/error, so this stays at info.
// The query string is not logged: it may carry tokens or personal search terms.
func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:   true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURIPath:   true,
		LogStatus:    true,
		LogRequestID: true,
		// Run the error handler first so the logged status is the final one.
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
