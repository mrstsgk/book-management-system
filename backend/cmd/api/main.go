package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mrstsgk/book-management-system/backend/config"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	httpcommon "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

// @title Book Management System API
// @version 0.1.0
// @BasePath /
func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// run keeps os.Exit out of the body so deferred cleanup (DB close) runs.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(cfg.NewLogger())

	db, err := pgcommon.Connect(pgcommon.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := pgcommon.Close(db); err != nil {
			slog.Error("failed to close db", "error", err)
		}
	}()

	e := httpcommon.NewEcho()

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}
