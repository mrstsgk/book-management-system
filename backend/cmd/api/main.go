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
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/catalog"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/googlebooks"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/openbd"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	pgtag "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/tag"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
	httpcommon "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

// @title Book Management System API
// @version 0.1.0
// @BasePath /
// @securityDefinitions.apikey AdminToken
// @in header
// @name Authorization
// @description 登録・更新・削除とカタログの確認に必要。「Bearer <ADMIN_TOKEN>」の形式で指定する
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

	bookCatalog := newCatalog(cfg.Catalog, &http.Client{Timeout: catalogTimeout})
	if err := seedIfEmpty(context.Background(), pgbook.NewRepository(db), pgbook.NewQuery(db), bookCatalog); err != nil {
		return err
	}

	e := httpcommon.NewEcho()
	registerRoutes(e, db, bookCatalog, cfg.AdminToken)

	// run はグレースフルシャットダウンの後に返るので、そこで書影の埋め直しも止める。
	// defer は後入れ先出しなので、この defer は DB クローズの defer より先（実行中の1冊が終わるまで待ってから）走る。
	ctx, cancel := context.WithCancel(context.Background())
	fillDone := startFillMissingCovers(ctx, &bookcmd.FillMissingCoversUsecaseImpl{Books: pgbook.NewRepository(db), Catalog: bookCatalog})
	defer func() {
		cancel()
		<-fillDone
	}()

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// startFillMissingCovers は起動時に1回だけ、書影の無い本の書影を埋める goroutine を起動してすぐ返す
// （外部カタログが遅くても起動を待たせないため）。DB のクローズと競合しないよう、呼び出し側は返り値の
// channel が閉じるまで待ってからクローズすること。
func startFillMissingCovers(ctx context.Context, uc bookcmd.FillMissingCoversUsecase) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := uc.Execute(ctx); err != nil {
			slog.WarnContext(ctx, "filling missing covers failed", "error", err)
		}
	}()
	return done
}

// catalogTimeout は外部カタログ1回の問い合わせの上限。遅い提供元にリクエストを長く占有させないため。
const catalogTimeout = 5 * time.Second

// newCatalog は書誌を openBD から取り、Google Books のキーがあれば書影を Google Books → openBD の順で探す。
func newCatalog(cfg config.CatalogConfig, client *http.Client) domainbook.BookCatalog {
	bibliography := openbd.NewCatalog(cfg.OpenBDBaseURL, client)
	if !cfg.GoogleBooksEnabled() {
		// nil の *googlebooks.Client を渡すと interface としては nil でなくなるので、ここで分ける
		return catalog.New(bibliography, nil)
	}
	return catalog.New(bibliography, googlebooks.NewClient(cfg.GoogleBooksBaseURL, cfg.GoogleBooksAPIKey, client))
}

// registerRoutes は手書きの DI（infra → usecase → presentation）で各ハンドラを組み立てて登録する。
func registerRoutes(e *echo.Echo, db *gorm.DB, bookCatalog domainbook.BookCatalog, adminToken string) {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db)
	tags := pgtag.NewRepository(db)
	tagQuery := pgtag.NewQuery(db)
	adminOnly := httpcommon.RequireAdminToken(adminToken)

	api := e.Group("/api")
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
		AdminOnly: adminOnly,
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
