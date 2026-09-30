package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/config"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	infraauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/openbd"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/auth"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
	httpauth "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/auth"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
	httpcommon "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/query"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

// @title Book Management System API
// @version 0.1.0
// @BasePath /
// @description 書き込み系 API は POST /api/auth/login で得た httpOnly Cookie（admin_session）が要る
func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// run は本体で os.Exit を呼ばないことで、defer の後始末（DB のクローズ）を必ず走らせる。
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

	bookCatalog := newCatalog(cfg.Catalog)
	if err := seedIfEmpty(context.Background(), pgbook.NewRepository(db), pgbook.NewQuery(db), bookCatalog); err != nil {
		return err
	}

	e := httpcommon.NewEcho()
	registerRoutes(e, db, bookCatalog, cfg)

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// catalogTimeout は外部カタログ1回の問い合わせの上限。遅い提供元にリクエストを長く占有させないため。
const catalogTimeout = 5 * time.Second

// newCatalog は書誌と書影を openBD から取る。
func newCatalog(cfg config.CatalogConfig) domainbook.BookCatalog {
	return openbd.NewCatalog(cfg.OpenBDBaseURL, &http.Client{Timeout: catalogTimeout})
}

// registerRoutes は手書きの DI（infra → usecase → presentation）で各ハンドラを組み立てて登録する。
func registerRoutes(e *echo.Echo, db *gorm.DB, bookCatalog domainbook.BookCatalog, cfg config.Config) {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)
	tags := pgtag.NewRepository(db)
	tagQuery := pgtag.NewQuery(db)
	sessions := pgauth.NewRepository(db)

	checkSession := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now}
	// 書き込み系はセッションの検証と Origin の検証を両方通す
	adminOnly := func(next echo.HandlerFunc) echo.HandlerFunc {
		return httpcommon.RequireSameOrigin()(httpcommon.RequireAdminSession(checkSession)(next))
	}

	api := e.Group("/api")
	(&httpauth.Handler{
		LoginUC: &authcmd.LoginUsecaseImpl{
			Admin:    authcmd.AdminAccount{ID: cfg.Admin.ID, PasswordHash: cfg.Admin.PasswordHash},
			Verifier: infraauth.NewBcryptVerifier(),
			Sessions: sessions,
			Now:      time.Now,
		},
		LogoutUC:   &authcmd.LogoutUsecaseImpl{Sessions: sessions},
		CheckUC:    checkSession,
		SameOrigin: httpcommon.RequireSameOrigin(),
	}).Register(api.Group("/auth"))
	(&httpbook.Handler{
		RegisterUC: &bookcmd.RegisterUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery},
		UpdateUC:   &bookcmd.UpdateUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery},
		DeleteUC:   &bookcmd.DeleteUsecaseImpl{Books: books},
		GetUC:      &bookqry.GetUsecaseImpl{Books: bookQuery},
		ListUC:     &bookqry.ListUsecaseImpl{Books: bookQuery},
		AdminOnly:  adminOnly,
	}).Register(api.Group("/books"))
	(&httpcatalog.Handler{
		LookupUC:  &bookqry.LookupCatalogUsecaseImpl{Catalog: bookCatalog},
		AdminOnly: httpcommon.RequireAdminSession(checkSession), // GET なので Origin 検証は要らない
	}).Register(api.Group("/catalog"))

	(&httptag.Handler{
		RegisterUC:   &tagcmd.RegisterUsecaseImpl{Tags: tags},
		RenameUC:     &tagcmd.RenameUsecaseImpl{Tags: tags},
		DeleteUC:     &tagcmd.DeleteUsecaseImpl{Tags: tags},
		ListUC:       &tagqry.ListUsecaseImpl{Tags: tagQuery},
		CountBooksUC: &tagqry.CountBooksUsecaseImpl{Tags: tagQuery},
		AdminOnly:    adminOnly,
	}).Register(api.Group("/tags"))
}
