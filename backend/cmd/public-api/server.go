package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"

	publicapi "github.com/mrstsgk/book-management-system/backend/api/gen/public"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	tagpublic "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/public"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// 各領域の Handler は型名がどれも Handler なので、埋め込んだときのフィールド名が重ならないよう別名を付ける。
type (
	bookPublicHandler = bookpublic.Handler
	tagPublicHandler  = tagpublic.Handler
)

// server は public チャネルの各領域の Handler を、生成された ServerInterface にまとめる。
// 領域を足すときはその Handler をここに埋め込む。実装していない operation があればコンパイルが通らない。
// 生成された IF 自体は埋め込まない（埋め込むと未実装が実行時エラーに化けて見つからなくなるため）。
type server struct {
	*bookPublicHandler
	*tagPublicHandler
}

var _ publicapi.ServerInterface = (*server)(nil)

// deps は buildEcho が組み立てに使う外側の依存（DB）。run は本物を、テストは Fake を渡す。
type deps struct {
	BookQuery bookdomain.Query
	TagQuery  tagdomain.Query
}

// buildEcho は入力検証・ルート登録・各領域の Handler を手書きの DI で組み立てる。
// run から分けているのは、DB 接続とポートの待ち受けなしで配線をテストできるようにするため。
// spec を引数にしているのは、入力検証の middleware が登録されていることをテストで確かめられるようにするため。
func buildEcho(spec *openapi3.T, d deps) (*echo.Echo, error) {
	validate, err := httpserver.ValidateRequest(spec)
	if err != nil {
		return nil, err
	}
	e := httpserver.NewEcho()
	publicapi.RegisterHandlersWithOptions(e, &server{
		bookPublicHandler: &bookpublic.Handler{
			ReadBookUsecase:    &bookqry.ReadBookUsecaseImpl{Books: d.BookQuery},
			SearchBooksUsecase: &bookqry.SearchBooksUsecaseImpl{Books: d.BookQuery},
		},
		tagPublicHandler: &tagpublic.Handler{
			SearchTagsUsecase: &tagqry.SearchTagsUsecaseImpl{Tags: d.TagQuery},
			CountBooksUsecase: &tagqry.CountBooksUsecaseImpl{Tags: d.TagQuery},
		},
	}, publicapi.RegisterHandlersOptions{
		// public の spec にはセッションを要る operation が無いので、セッションの middleware は渡さない
		// （渡す相手が無い。security を足したら nil の middleware でリクエストが落ちて気づける）
		OperationMiddlewares: httpserver.OperationMiddlewares(spec, httpserver.RequireSameOrigin(), nil, validate),
	})
	return e, nil
}
