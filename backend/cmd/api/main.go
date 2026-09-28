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
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/openbd"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/rakuten"
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

	bookCatalog := newCatalog(cfg.Catalog)
	if err := seedIfEmpty(context.Background(), pgbook.NewRepository(db), pgbook.NewQuery(db, time.Now), bookCatalog); err != nil {
		return err
	}

	e := httpcommon.NewEcho()
	registerRoutes(e, db, bookCatalog, cfg.AdminToken)

	// run はグレースフルシャットダウンの後に返るので、そこで取り直しも止める。
	// defer は後入れ先出しなので、この defer は DB クローズの defer より先（実行中の1回が終わるまで待ってから）走る。
	ctx, cancel := context.WithCancel(context.Background())
	refreshDone := startRakutenRefresh(ctx, &bookcmd.RefreshRakutenCoversUsecaseImpl{
		Books: pgbook.NewRepository(db), Catalog: bookCatalog, Now: time.Now,
	}, rakutenRefreshInterval)
	defer func() {
		cancel()
		<-refreshDone
	}()

	return httpcommon.Serve(e, fmt.Sprintf(":%s", cfg.HTTPPort))
}

// rakutenRefreshInterval は楽天の書影を取り直す間隔（要件定義 §1.2「稼働中は1日1回」）。
const rakutenRefreshInterval = 24 * time.Hour

// startRakutenRefresh は起動直後に1回、その後 interval ごとに楽天の書影を取り直す goroutine を起動してすぐ返す。
// 外部カタログが遅くても起動を待たせないため、別の goroutine で回す。ctx が終わっても、実行中の1回は
// 中断せず最後まで走らせてから止まる（DB のクローズと競合しないよう、呼び出し側は返り値の channel が
// 閉じるまで待ってからクローズすること）。
func startRakutenRefresh(ctx context.Context, uc bookcmd.RefreshRakutenCoversUsecase, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := uc.Execute(ctx); err != nil {
				slog.WarnContext(ctx, "rakuten cover refresh failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return done
}

// catalogTimeout は外部カタログ1回の問い合わせの上限。遅い提供元にリクエストを長く占有させないため。
const catalogTimeout = 5 * time.Second

// newCatalog は openBD を使い、楽天のキーが設定されていれば書影の無い本を楽天ブックスで補う。
func newCatalog(cfg config.CatalogConfig) domainbook.BookCatalog {
	client := &http.Client{Timeout: catalogTimeout}
	primary := openbd.NewCatalog(cfg.OpenBDBaseURL, client)
	var fallback domainbook.BookCatalog
	if cfg.RakutenEnabled() {
		fallback = rakuten.NewCatalog(cfg.RakutenBaseURL, cfg.RakutenApplicationID, cfg.RakutenAccessKey, client, time.Now)
	} else {
		slog.Info("RAKUTEN_APPLICATION_ID / RAKUTEN_ACCESS_KEY not set; covers come from openBD only")
	}
	return catalog.NewChain(primary, fallback)
}

// registerRoutes は手書きの DI（infra → usecase → presentation）で各ハンドラを組み立てて登録する。
func registerRoutes(e *echo.Echo, db *gorm.DB, bookCatalog domainbook.BookCatalog, adminToken string) {
	books := pgbook.NewRepository(db)
	bookQuery := pgbook.NewQuery(db, time.Now)
	tags := pgtag.NewRepository(db)
	tagQuery := pgtag.NewQuery(db)
	adminOnly := httpcommon.RequireAdminToken(adminToken)

	api := e.Group("/api")
	(&httpbook.Handler{
		RegisterUC: &bookcmd.RegisterUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery},
		UpdateUC:   &bookcmd.UpdateUsecaseImpl{Books: books, Catalog: bookCatalog, Details: bookQuery, Tags: tagQuery, Now: time.Now},
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
