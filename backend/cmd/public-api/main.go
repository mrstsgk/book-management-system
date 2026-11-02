package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	"github.com/mrstsgk/book-management-system/backend/config"
	authbcrypt "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/bcrypt"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/auth/infrastructure/postgres"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/gateway/openbd"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/tag/infrastructure/postgres"
)

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

	spec, err := webapi.GetSpec()
	if err != nil {
		return err
	}
	e, err := buildEcho(spec, deps{
		Books:     pgbook.NewRepository(db),
		BookQuery: pgbook.NewQuery(db),
		Catalog:   bookCatalog,
		Tags:      pgtag.NewRepository(db),
		TagQuery:  pgtag.NewQuery(db),
		Sessions:  pgauth.NewRepository(db),
		Admin:     authcmd.AdminAccount{ID: cfg.Admin.ID, PasswordHash: cfg.Admin.PasswordHash},
		Verifier:  authbcrypt.NewBcryptVerifier(),
		Now:       time.Now,
	})
	if err != nil {
		return err
	}

	return httpserver.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// catalogTimeout は外部カタログ1回の問い合わせの上限。遅い提供元にリクエストを長く占有させないため。
const catalogTimeout = 5 * time.Second

// newCatalog は書誌と書影を openBD から取る。
func newCatalog(cfg config.CatalogConfig) bookdomain.BookCatalog {
	return openbd.NewCatalog(cfg.OpenBDBaseURL, &http.Client{Timeout: catalogTimeout})
}
