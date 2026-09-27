package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/config"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	gwbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/book"
	pgauthor "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/author"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	httpauthor "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/author"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	httpcommon "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	authorcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
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

	images, err := gwbook.NewImageStorage(context.Background(), gwbook.Config{
		Endpoint:        cfg.S3.Endpoint,
		Region:          cfg.S3.Region,
		Bucket:          cfg.S3.Bucket,
		AccessKeyID:     cfg.S3.AccessKeyID,
		SecretAccessKey: cfg.S3.SecretAccessKey,
	})
	if err != nil {
		return err
	}

	e := httpcommon.NewEcho()
	registerRoutes(e, db, images)

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// registerRoutes is the hand-written DI: infra → usecase → presentation.
func registerRoutes(e *echo.Echo, db *gorm.DB, images domainbook.ImageStorage) {
	authorRepo := pgauthor.NewRepository(db)
	authorQuery := pgauthor.NewQuery(db)
	bookRepo := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)

	api := e.Group("/api")
	(&httpauthor.Handler{
		CreateUC: &authorcmd.CreateUsecaseImpl{Authors: authorRepo},
		UpdateUC: &authorcmd.UpdateUsecaseImpl{Authors: authorRepo},
		BooksUC:  &bookqry.ListByAuthorUsecaseImpl{Authors: authorQuery, Books: bookQuery},
	}).Register(api.Group("/authors"))
	(&httpbook.Handler{
		CreateUC:      &bookcmd.CreateUsecaseImpl{Books: bookRepo, Authors: authorRepo, Details: bookQuery, Images: images},
		UpdateUC:      &bookcmd.UpdateUsecaseImpl{Books: bookRepo, Authors: authorRepo, Details: bookQuery, Images: images},
		UploadImageUC: &bookcmd.UploadImageUsecaseImpl{Books: bookRepo, Images: images, Details: bookQuery},
		GetUC:         &bookqry.GetUsecaseImpl{Books: bookQuery, Images: images},
	}).Register(api.Group("/books"))
}
