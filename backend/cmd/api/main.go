package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/config"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/catalog"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/openbd"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/rakuten"
	pgauthor "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/author"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	httpauthor "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/author"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
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

	catalog := newCatalog(cfg.Catalog)

	e := httpcommon.NewEcho()
	registerRoutes(e, db, catalog)

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// registerRoutes is the hand-written DI: infra → usecase → presentation.
// catalogTimeout bounds each external catalog call so a slow provider can't hold a request.
const catalogTimeout = 5 * time.Second

// newCatalog uses openBD and, when its keys are set, Rakuten Books to fill missing covers.
func newCatalog(cfg config.CatalogConfig) domainbook.BookCatalog {
	client := &http.Client{Timeout: catalogTimeout}
	primary := openbd.NewCatalog(cfg.OpenBDBaseURL, client)
	var fallback domainbook.BookCatalog
	if cfg.RakutenEnabled() {
		fallback = rakuten.NewCatalog(cfg.RakutenBaseURL, cfg.RakutenApplicationID, cfg.RakutenAccessKey, client)
	} else {
		slog.Info("RAKUTEN_APPLICATION_ID / RAKUTEN_ACCESS_KEY not set; covers come from openBD only")
	}
	return catalog.NewChain(primary, fallback)
}

func registerRoutes(e *echo.Echo, db *gorm.DB, bookCatalog domainbook.BookCatalog) {
	authorRepo := pgauthor.NewRepository(db)
	authorQuery := pgauthor.NewQuery(db)
	bookRepo := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)
	bookListQuery := pgbook.NewListQuery(db)

	api := e.Group("/api")
	(&httpauthor.Handler{
		CreateUC: &authorcmd.CreateUsecaseImpl{Authors: authorRepo},
		UpdateUC: &authorcmd.UpdateUsecaseImpl{Authors: authorRepo},
		BooksUC:  &bookqry.ListByAuthorUsecaseImpl{Authors: authorQuery, Books: bookQuery},
	}).Register(api.Group("/authors"))
	(&httpbook.Handler{
		CreateUC: &bookcmd.CreateUsecaseImpl{Books: bookRepo, Authors: authorRepo, Details: bookQuery, Catalog: bookCatalog},
		UpdateUC: &bookcmd.UpdateUsecaseImpl{Books: bookRepo, Authors: authorRepo, Details: bookQuery, Catalog: bookCatalog},
		GetUC:    &bookqry.GetUsecaseImpl{Books: bookQuery},
		ListUC:   &bookqry.ListUsecaseImpl{Books: bookListQuery},
	}).Register(api.Group("/books"))
	(&httpcatalog.Handler{
		LookupUC: &bookqry.LookupCatalogUsecaseImpl{Catalog: bookCatalog},
	}).Register(api.Group("/catalog"))
}
