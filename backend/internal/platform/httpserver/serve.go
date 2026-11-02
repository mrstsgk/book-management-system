package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

// shutdownTimeout は処理中のリクエストを待つ上限。オーケストレータの停止猶予（一般に 30 秒）より短くしておく。
const shutdownTimeout = 10 * time.Second

// Serve は SIGINT / SIGTERM まで e を動かし、その後 graceful shutdown する。
func Serve(e *echo.Echo, addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server started", "addr", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return err
	}
	slog.Info("graceful shutdown completed")
	return nil
}
