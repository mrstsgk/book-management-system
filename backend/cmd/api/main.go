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
	authbcrypt "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/bcrypt"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/postgres"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/gateway/openbd"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	httpauth "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/auth"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
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

	db, err := pgplatform.Connect(pgplatform.Config{
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
		if err := pgplatform.Close(db); err != nil {
			slog.Error("failed to close db", "error", err)
		}
	}()

	bookCatalog := newCatalog(cfg.Catalog)
	if err := seedIfEmpty(context.Background(), pgbook.NewRepository(db), pgbook.NewQuery(db), bookCatalog); err != nil {
		return err
	}

	e := httpserver.NewEcho()
	registerRoutes(e, db, bookCatalog, cfg)

	return httpserver.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// catalogTimeout は外部カタログ1回の問い合わせの上限。遅い提供元にリクエストを長く占有させないため。
const catalogTimeout = 5 * time.Second

// newCatalog は書誌と書影を openBD から取る。
func newCatalog(cfg config.CatalogConfig) bookdomain.BookCatalog {
	return openbd.NewCatalog(cfg.OpenBDBaseURL, &http.Client{Timeout: catalogTimeout})
}

// registerRoutes は手書きの DI（infra → usecase → presentation）で各ハンドラを組み立てて登録する。
func registerRoutes(e *echo.Echo, db *gorm.DB, bookCatalog bookdomain.BookCatalog, cfg config.Config) {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)
	tags := pgtag.NewRepository(db)
	tagQuery := pgtag.NewQuery(db)
	sessions := pgauth.NewRepository(db)

	checkSession := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now}
	// 書き込み系はセッションの検証と Origin の検証を両方通す
	adminOnly := func(next echo.HandlerFunc) echo.HandlerFunc {
		return httpserver.RequireSameOrigin()(httpserver.RequireAdminSession(checkSession)(next))
	}

	api := e.Group("/api")
	(&httpauth.Handler{
		LoginUsecase: &authcmd.LoginUsecaseImpl{
			Admin:    authcmd.AdminAccount{ID: cfg.Admin.ID, PasswordHash: cfg.Admin.PasswordHash},
			Verifier: authbcrypt.NewBcryptVerifier(),
			Sessions: sessions,
			Now:      time.Now,
		},
		LogoutUsecase:       &authcmd.LogoutUsecaseImpl{Sessions: sessions},
		CheckSessionUsecase: checkSession,
		SameOrigin:          httpserver.RequireSameOrigin(),
	}).Register(api.Group("/auth"))
	(&httpbook.Handler{
		RegisterBookUsecase: &bookcmd.RegisterBookUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery},
		UpdateBookUsecase:   &bookcmd.UpdateBookUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery},
		DeleteBookUsecase:   &bookcmd.DeleteBookUsecaseImpl{Books: books},
		ReadBookUsecase:     &bookqry.ReadBookUsecaseImpl{Books: bookQuery},
		SearchBooksUsecase:  &bookqry.SearchBooksUsecaseImpl{Books: bookQuery},
		AdminOnly:           adminOnly,
	}).Register(api.Group("/books"))
	(&httpcatalog.Handler{
		LookupCatalogUsecase: &bookqry.LookupCatalogUsecaseImpl{Catalog: bookCatalog},
		AdminOnly:            httpserver.RequireAdminSession(checkSession), // GET なので Origin 検証は要らない
	}).Register(api.Group("/catalog"))

	(&httptag.Handler{
		RegisterTagUsecase: &tagcmd.RegisterTagUsecaseImpl{Tags: tags, TagQuery: tagQuery},
		RenameTagUsecase:   &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: tagQuery},
		DeleteTagUsecase:   &tagcmd.DeleteTagUsecaseImpl{Tags: tags},
		SearchTagsUsecase:  &tagqry.SearchTagsUsecaseImpl{Tags: tagQuery},
		CountBooksUsecase:  &tagqry.CountBooksUsecaseImpl{Tags: tagQuery},
		AdminOnly:          adminOnly,
	}).Register(api.Group("/tags"))
}
