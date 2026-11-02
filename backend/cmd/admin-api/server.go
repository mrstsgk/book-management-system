package main

import (
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"

	adminapi "github.com/mrstsgk/book-management-system/backend/api/gen/admin"
	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
	authadmin "github.com/mrstsgk/book-management-system/backend/internal/auth/presentation/http/admin"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookadmin "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/admin"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	tagadmin "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/admin"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// 各領域の Handler は型名がどれも Handler なので、埋め込んだときのフィールド名が重ならないよう別名を付ける。
type (
	bookAdminHandler = bookadmin.Handler
	catalogHandler   = bookadmin.CatalogHandler
	tagAdminHandler  = tagadmin.Handler
	authAdminHandler = authadmin.Handler
)

// server は admin チャネルの各領域の Handler を、生成された ServerInterface にまとめる。
// 領域を足すときはその Handler をここに埋め込む。実装していない operation があればコンパイルが通らない。
// 生成された IF 自体は埋め込まない（埋め込むと未実装が実行時エラーに化けて見つからなくなるため）。
type server struct {
	*bookAdminHandler
	*catalogHandler
	*tagAdminHandler
	*authAdminHandler
}

var _ adminapi.ServerInterface = (*server)(nil)

// deps は buildEcho が組み立てに使う外側の依存（DB・外部カタログ・設定）。
// run は本物を、テストは Fake を渡す。
type deps struct {
	Books     bookdomain.Repository
	BookQuery bookdomain.Query
	Catalog   bookdomain.BookCatalog
	Tags      tagdomain.Repository
	TagQuery  tagdomain.Query
	Sessions  authdomain.SessionRepository
	Admin     authcmd.AdminAccount
	Verifier  authdomain.PasswordVerifier
	Now       func() time.Time
}

// buildEcho は入力検証・ルート登録・認証（operation ごとの middleware）・各領域の Handler を手書きの DI で組み立てる。
// run から分けているのは、DB 接続とポートの待ち受けなしで配線をテストできるようにするため。
// spec を引数にしているのは、入力検証の middleware が登録されていることをテストで確かめられるようにするため。
func buildEcho(spec *openapi3.T, d deps) (*echo.Echo, error) {
	validate, err := httpserver.ValidateRequest(spec)
	if err != nil {
		return nil, err
	}
	checkSession := &authqry.CheckSessionUsecaseImpl{Sessions: d.Sessions, Now: d.Now}

	e := httpserver.NewEcho()
	adminapi.RegisterHandlersWithOptions(e, &server{
		bookAdminHandler: &bookadmin.Handler{
			ReadBookUsecase:     &bookqry.ReadBookUsecaseImpl{Books: d.BookQuery},
			SearchBooksUsecase:  &bookqry.SearchBooksUsecaseImpl{Books: d.BookQuery},
			RegisterBookUsecase: &bookcmd.RegisterBookUsecaseImpl{Books: d.Books, Catalog: d.Catalog, Details: d.BookQuery, Tags: d.TagQuery},
			UpdateBookUsecase:   &bookcmd.UpdateBookUsecaseImpl{Books: d.Books, Catalog: d.Catalog, Details: d.BookQuery, Tags: d.TagQuery},
			DeleteBookUsecase:   &bookcmd.DeleteBookUsecaseImpl{Books: d.Books},
		},
		catalogHandler: &bookadmin.CatalogHandler{
			LookupCatalogUsecase: &bookqry.LookupCatalogUsecaseImpl{Catalog: d.Catalog},
		},
		tagAdminHandler: &tagadmin.Handler{
			SearchTagsUsecase:  &tagqry.SearchTagsUsecaseImpl{Tags: d.TagQuery},
			CountBooksUsecase:  &tagqry.CountBooksUsecaseImpl{Tags: d.TagQuery},
			RegisterTagUsecase: &tagcmd.RegisterTagUsecaseImpl{Tags: d.Tags, TagQuery: d.TagQuery},
			RenameTagUsecase:   &tagcmd.RenameTagUsecaseImpl{Tags: d.Tags, TagQuery: d.TagQuery},
			DeleteTagUsecase:   &tagcmd.DeleteTagUsecaseImpl{Tags: d.Tags},
		},
		authAdminHandler: &authadmin.Handler{
			LoginUsecase:  &authcmd.LoginUsecaseImpl{Admin: d.Admin, Verifier: d.Verifier, Sessions: d.Sessions, Now: d.Now},
			LogoutUsecase: &authcmd.LogoutUsecaseImpl{Sessions: d.Sessions},
		},
	}, adminapi.RegisterHandlersOptions{
		OperationMiddlewares: httpserver.OperationMiddlewares(spec, httpserver.RequireSameOrigin(), httpserver.RequireAdminSession(checkSession), validate),
	})
	return e, nil
}
