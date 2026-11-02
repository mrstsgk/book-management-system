# バックエンド方針

## 文書の分担（MECE）

| 文書 | 管轄 |
|---|---|
| [`docs/architecture.md`](../docs/architecture.md) | リポジトリ全体・トップレベル構成・OpenAPI／FE との契約の境界 |
| **本書** | バックエンドのスタック・層・HTTP/OpenAPI・ディレクトリ・永続化 |
| [`frontend/architecture.md`](../frontend/architecture.md) | フロントのスタック・設計。OpenAPI **生成物の利用**のみ |

人が読む方針は **日本語**（AI 入口は `CLAUDE.md` / `.cursor/rules`）。

## 1. 決定（要約）

| 項目 | 決定 |
|---|---|
| 形 | **Go モノリス**（1 `go.mod`）。API バイナリは `cmd/api` の1つ |
| HTTP | **Echo** |
| 設計 | **オニオン** + DDD 戦術 + **CQRS**（単一 DB。全体 Event Sourcing はしない） |
| 永続化 | **PostgreSQL + GORM**（現在状態が正）。**GORM AutoMigrate は使わない** |
| マイグレーション | **golang-migrate**（`backend/migrations/` が SQL の正。アプリ起動時 migrate しない） |
| HTTP / OpenAPI | Echo + validator + swag。**Go の DTO／Handler が BE の正** |
| FE 契約 | swag **排出 OpenAPI → TypeScript 生成は必須**（手編集禁止・CI ドリフト検知） |
| 書誌・書影 | 外部カタログ。Domain の `book.BookCatalog`（ExternalGateway）を `infrastructure/gateway/openbd` が実装する。画像は保存しない。openBD に書影が無ければ書影なし（[ADR](../docs/adr/2026-09-30-cover-from-openbd-only.md)） |
| 認証 | 書き込み系はログイン済みのセッション Cookie（`presentation/http/common.RequireAdminSession` + `RequireSameOrigin`）。閲覧は認証なし |
| ツールチェーン | **mise で Go 版を固定**（リポジトリ直下 `.mise.toml`。`go.mod` と揃える） |
| ローカル開発 | API は**ホストの Go**、DB は **Docker Compose**。Dev Container なし |
| 旧スタック | Kotlin / Spring Boot / jOOQ / Flyway と、旧仕様（書籍・著者の管理 API）の Go 版は**削除済み**（[ADR](../docs/adr/2026-09-28-rebuild-as-reading-portfolio.md)） |

マイクロサービス分割はしない。

### 1.1 ローカル開発

| 対象 | どう動かすか |
|---|---|
| Go 本体 | `mise install`（`.mise.toml` の `go`）→ ホストで `go run` / `make run` |
| PostgreSQL | `docker compose up -d`（`backend/docker-compose.yml`） |
| migrate | ホストの `golang-migrate`（`make migrate-up`）。アプリ起動時には走らせない |

Dev Container で IDE ごとコンテナに閉じ込める方式は採らない。

## 2. 層と依存

```text
presentation/http/{common,<domain>}
  → usecase/<domain>/{command,query}
    → domain/{common,<domain>}
       ↑
infrastructure/postgres/{common,<domain>}, infrastructure/gateway/{<domain>}
```

各層の直下は **横断 common + ドメイン別ディレクトリ** で揃える（フラットに全ファイルを並べない）。

| 層 | 責務 |
|---|---|
| **Presentation**（`presentation/http`） | HTTP リクエスト／レスポンス変換のみ。`common`（Echo・`BindValidate`・`HTTPErrorHandler`）とドメイン別 Handler／DTO。**ビジネスロジックも認可も持たない** |
| **UseCase**（`usecase/<domain>`） | シナリオ実行。**認可**（必要なとき）と業務フロー。CQRS はドメイン内の `command` / `query` |
| **Domain**（`domain`） | ビジネスルールの中核。`common`（sentinel error）とドメイン別 Entity／VO／**Repository IF**／**Query IF**／**ExternalGateway IF**（外部 API・SaaS 連携の抽象）。**他層に依存しない** |
| **Infrastructure**（`infrastructure/postgres`, `infrastructure/gateway`） | `postgres/<domain>` が Repository／Query 実装。`gateway/<domain>` が ExternalGateway 実装（外部 API・SaaS 呼び出し） |

- Domain Service はエンティティに載せにくい業務ルール
- 命名目安: UseCase は IF `RegisterBookUsecase` + 公開の実装 `RegisterBookUsecaseImpl`、Handler は `Handler`（詳細は下の「配置と命名のルール」「型・メソッドの名前の付け方」）
- `cmd/api/main.go` で手書き DI（コンテナなし）: config → infra → usecase → presentation（`registerRoutes`）
- 外部 API・SaaS 呼び出しは UseCase から **ExternalGateway IF**（Domain 定義）経由で行う。Repository／Query と同じ理由（Domain が具体的な HTTP クライアント実装に依存しないため）で Domain に IF を置く

### 配置と命名のルール（迷ったらここを写す）

- Go のパッケージ名はディレクトリ名と一致させる（`book` / `query` / `command` / `common`）
- 複数の層・ドメインを import する側（`cmd/api` 等）は `<役割><ドメイン>` の別名を付ける。実例（`cmd/api/main.go`）: `domainbook`・`pgbook`・`pgcommon`・`httpbook`・`httpcommon`・`bookcmd`・`bookqry`・`authcmd`・`tagqry`
- **`Impl` は IF と実装が同じパッケージにあるときだけ付ける**（命名に迷わず全ドメインで同じ形にするため、条件を 1 つに決めておく）
  - UseCase: 同じパッケージに IF `XxxUsecase` と公開 struct `XxxUsecaseImpl` を置き、`cmd` で struct リテラル（`&bookcmd.RegisterUsecaseImpl{Books: ..., Catalog: ...}`）から組み立てる。Handler は IF にだけ依存するので、実装を変えても Handler は変わらない
  - Infrastructure: IF は Domain にあるので名前は衝突しない。非公開 struct（`repository` / `query`）+ Domain の IF を返す `NewRepository(db)` / `NewQuery(db)` で作り、`Impl` は付けない。呼び出し側は Domain の IF にしか依存できない
- ファイル名は小文字（区切りは `_`）。ドメイン名や `_impl` を付けない（ディレクトリで分かる）。`_linux` / `_js` 等の OS・アーキテクチャ名で終わらせない（Go のビルド制約として解釈される）。テストは `_test.go` にし、テスト以外のファイルを `_test` で終わらせない。例: `register.go` / `register_test.go` / `query.go` / `repository.go`
- 会話・レビューでは同名を区別するためパッケージ名付きで呼ぶ（`book.Repository` = IF、`pgbook.NewRepository` = 実装、`bookcmd.RegisterUsecase` = UseCase）
- テストは対象と同じディレクトリに置く（[テスト規約](../docs/rules/testing.md)）
- **新しいドメインの追加・境界の変更は設計レビューの対象**。要件と実際の変更理由から決め、予測で空パッケージを量産しない。ディレクトリは `book` を写し、今必要な層だけ作る

### 型・メソッドの名前の付け方

名前は下の「設計原則」の目的駆動名前設計に従い、技術的な手段ではなく業務上の目的・振る舞いから付ける。用語は要件定義（[`docs/specifications.md`](../docs/specifications.md)）と同じ言葉を使う。

**UseCase**

- 型名は `<動詞><業務対象>Usecase`（綴りは `Usecase` のまま。`UseCase` にしない）。パッケージ名が `query` / `command` でドメインが呼び出し側から見えないため、型名に業務対象（`Book` / `Tag`）を入れる
- 単純な参照・更新・削除は `Read` / `Update` / `Delete`。登録・改名・ログイン・カタログの確認など業務上の意味を持つ操作は業務の動詞（`Register` / `Rename` / `Login` / `Logout` / `LookupCatalog`）にし、CRUD に寄せない
- `Read` は 1 件の参照だけに使う。一覧は、利用者が条件を指定するかどうかに関係なく `Search` にし、業務対象を複数形にする（`SearchBooksUsecase`）。`List` や `Get` は使わない（1 件と一覧が末尾の `s` だけで分かれて取り違えやすく、一覧の動詞を 1 つにして迷わないため）。複数形は複数件を返すことだけを表し、全件を返すとは限らない
- UseCase の名前を Repository／Query のメソッド名から決めない
- ファイル名は型名から `Usecase` を除いた snake_case（`search_books.go` / `rename_tag.go`）

| 業務の目的 | UseCase の型名 |
|---|---|
| 読んだ本を 1 件参照する | `ReadBookUsecase` |
| 読んだ本の一覧を参照する（検索・絞り込みを含む） | `SearchBooksUsecase` |
| 読んだ本を登録する | `RegisterBookUsecase` |
| 読んだ本を更新・削除する | `UpdateBookUsecase` / `DeleteBookUsecase` |
| 分野タグの名前を変える | `RenameTagUsecase` |
| 分野タグごとの冊数を数える | `CountBooksUsecase` |
| ログインする | `LoginUsecase` |

> **移行中**: 既存の `ListUsecase` / `GetUsecase` / `RegisterUsecase` / `UpdateUsecase` / `DeleteUsecase` / `RenameUsecase` はこの規則の前に付けた名前。改名は別の PR で行う（[ADR](../docs/adr/2026-11-02-revise-development-rules.md)）。新しく作るものはこの規則で付ける。

**Repository／Query のメソッド**

取得は `Find` を基本にし、目的が違うときだけ `Search` / `Count` / `Exists` を使う。同じ意味の取得に `Read` / `Get` / `Find` を混ぜない（UseCase の `Read` は業務の目的、Repository／Query の `Find` はデータの取得で、層の責務が違うため名前も違う）。更新系は `Create` / `Update` / `Delete`。

| 目的 | 名前 | 使う条件 |
|---|---|---|
| 一意なキーで 1 件 | `FindByID` | 条件が一意のときだけ。一意でない条件で 1 件を暗黙に選ばない |
| 全件 | `FindAll` | 全件を返す契約のときだけ |
| 固定の条件に合うものを全件 | `FindAllByTagID` | 一致したものを**すべて**返すときだけ。ページングした一部には使わない |
| 利用者が変えられる条件での検索 | `Search(criteria, range)` | キーワード・絞り込み・ページングなど。`WHERE` があるだけでは使わない |
| 件数 | `CountByTag` | 数値を返す |
| 存在するか | `ExistsAll` | `bool` を返す |

Query（読み取り側）は Read Model を返し、業務の状態を変えない。現状の `FindList`（`book.Query` / `tag.Query`）は条件付き・範囲付きの取得なので `Search` が規則どおりの名前、`auth.SessionRepository` の `Find` は `FindByID`、`Save` は `Create` に当たる。これらも上の移行中の改名に含める。

**Domain**

- Entity・集約・VO は業務の概念で名付け（`Book` / `Tag` / `ISBN` / `Review`）、`BookData` / `TagEntity` のような技術のサフィックスを付けない。種類が違うだけで Entity を機械的に分けない（業務の目的・不変条件・ライフサイクル・変更理由で判断する）
- メソッドは業務の振る舞いにする（`book.RefreshCatalog()` / `tag.Rename()`）。`UpdateStatus()` / `Process()` に業務の意味を隠さない。不変条件は状態を変えるメソッドの中で守る（事前に `Validate` を呼ばないと安全にならない形にしない）
- 判定は `bool` を返す `CanXxx`（操作できるか）/ `IsXxx`（状態）/ `HasXxx`（持っているか）。VO の生成条件は `NewXxx` がエラーを返して保証し、すべての型に `Validate` を付けない

### 層ごとの規則一覧（決まっていること）

層ごとに散らばっている規則の索引。正は各節（命名は上、HTTP は §3、テストは [テスト規約](../docs/rules/testing.md)、コメントは [コメントの書き方](../docs/rules/code-comments.md)）。「実例」は現在のコードにあるもの。

| 層 | 置き場・ファイル名 | 型・関数の命名 | テスト |
|---|---|---|---|
| **Domain** | `internal/domain/<domain>/`。Entity と IF は `<domain>.go`、VO は概念ごとのファイル（実例: `book/{book,isbn,review,summary,title}.go`、`tag/{tag,name}.go`、`auth/{session,credentials}.go`、`common/{error,list_range}.go`） | 書き込み IF `Repository`・読み取り IF `Query`・外部連携 IF は業務の能力で名付ける（実例: `book.BookCatalog`、`auth.PasswordVerifier`）。Read Model は `XxxDetail` / `XxxList` / `XxxListItem` / `XxxSummary` など用途が分かる名前（実例: `BookDetail`、`TagListItem`）。VO の不変条件は `NewXxx` で保証 | DB・HTTP に依存しない。不変条件・業務ルール・境界値と、不正な状態を作れないこと |
| **UseCase（参照）** | `internal/usecase/<domain>/query/`。1 UseCase 1 ファイル（実例: `book/query/{get,list,lookup_catalog}.go`、`tag/query/{list,count_books}.go`） | IF `XxxUsecase` + 公開 struct `XxxUsecaseImpl`、メソッドは `Execute`、依存は公開フィールド（`Books book.Query`）。`cmd` で struct リテラルから組み立てる | 手書き Fake を注入し DB 無しで動かす。下位層のエラーが伝播すること |
| **UseCase（更新）** | `internal/usecase/<domain>/command/`（実例: `book/command/{register,update,delete}.go`、`auth/command/{login,logout}.go`） | 参照と同じ。データを持つ入力は Command DTO `RegisterCommand` / `UpdateCommand`、ID だけなら scalar 引数 | 参照と同じに加え、失敗時に状態が変わらないこと、入力エラー時に下位層が呼ばれないこと |
| **Infrastructure（DB）** | `internal/infrastructure/postgres/<domain>/`。実装は `repository.go` / `query.go`、Persistence Model（`gorm` タグ）は同じファイルの非公開 `model`（実例: `book/{repository,query}.go`）。DB 接続は `postgres/common` | 非公開 struct（`repository` / `query`）+ Domain の IF を返す `NewRepository` / `NewQuery`。`Impl` は付けない。**型と各メソッドに日本語で概要の 1 行コメント** | 実際の PostgreSQL への契約テスト（繋がらなければ失敗）。SQL が複雑なもの（JOIN・集計・トランザクション）を優先し、Query は Read Model への詰め替えと書き込み側の型に依存しないことも確認 |
| **Infrastructure（外部連携）** | `internal/infrastructure/gateway/<提供元>/`（実例: `gateway/openbd/catalog.go`）。ハッシュ照合は `infrastructure/auth/bcrypt.go` | Domain の IF（`book.BookCatalog` 等）を実装。実装には提供元の名前を使ってよい | 偽の HTTP サーバ（`httptest`）に対する契約テスト |
| **Presentation** | `internal/presentation/http/<domain>/handler.go`。Request／Response DTO（`validate` タグ・swag コメント）も同じファイル（実例: `book/handler.go` の `RegisterRequest` / `Response` / `ListResponse`）。`presentation/http/common` = Echo・`BindValidate`・`HTTPErrorHandler`・セッション／同一オリジンの middleware | struct `Handler`、`Register(g *echo.Group)` でルート登録、メソッドは UseCase と同じ目的の名前。依存は公開フィールドで UseCase の IF | `handler_test.go`（Fake の UseCase で 400 と domain error → ステータスの変換）と `handler_integration_test.go`（実 DB で、返しうるステータスごとに 1 件） |
| **cmd** | `cmd/api/main.go`（設定・DB 接続・起動・`registerRoutes`）、`sample_data.go`（見本データ） | import の別名は `<役割><ドメイン>`（上の「配置と命名のルール」） | `main_test.go`（in-package）で配線（依存の nil・middleware の適用）を実リクエストで検証 |
| **全層共通** | ファイル名は小文字・`_` 区切り。ドメイン名・`_impl` を付けない | パッケージ名 = ディレクトリ名。会話ではパッケージ名付きで呼ぶ | 対象と同じディレクトリ。原則は外部テストパッケージ `<対象>_test`。入力違いはテーブルドリブン。DB・環境変数・グローバル状態を共有しないテストには `t.Parallel()` |

### 未決の規則（初めて作るときに決める）

まだ実例が無いか、現在のコードで書き方が揃っていないもの。該当する実装が初めて出る PR で決め、決めたら上の一覧と各節に反映する。推測で先に決めない。

| 対象 | 決めること |
|---|---|
| ID の型 | 今は `book.ID` / `tag.ID` の型（`int64` の別名）。VO にするか |
| Fake の置き場 | 各テストファイルに書くか、共有のテスト用パッケージを作るか（今は各ファイル） |
| テスト関数名 | `TestRepository_FindByID_NotFound` 形式（英語）と `TestXxx_日本語の振る舞い` の混在をどちらに揃えるか |
| DB を使うテストの並列実行 | 接続先・片付けは決定済み（[テスト規約](../docs/rules/testing.md)）。`t.Parallel()` を付けるか |

### 設計原則（目的駆動）

- **目的駆動名前設計**: 名前は「存在」ではなく「目的」で付ける。1 手段に対して目的は 1 つ
- **1 クラス 1 目的**: 目的が増えるほど変更の影響が広がる。目的ごとに分離する（例: 書籍の新規登録と更新で入力・検証が異なるなら `CreateCommand` / `UpdateCommand` と UseCase を分ける）
- 共通化より「変更理由の分離」を優先する（似ているだけの処理を無理に1つへまとめない。「同じ理由で変更されるか？」で判断し、`Common` / `Util` / `Manager` / `Helper` / `BaseService` のような曖昧な共通クラスを安易に作らない）
- **Domain は外部技術に依存しない**（§2 表の「他層に依存しない」の具体化）: Web Framework（Echo）／Controller／Request・Response／ORM（GORM）／SQL／DB 固有型／外部 API SDK／Infrastructure 実装のいずれも Domain から直接参照しない。外部との通信が要るときは Repository／Query／ExternalGateway の Interface を経由する
- **業務ルールは Domain（VO・Entity）に閉じ込め、不正な状態を生成できない設計を優先する**: 呼び出し側の複数箇所に同じ範囲チェックを書く（例: `if price >= 0 && price <= 99999999` があちこちに存在する）のではなく、VO のコンストラクタ1箇所で不変条件を保証する（例: 書籍価格 VO のコンストラクタが 0 以上を保証する）
- **モデルは目的ごとに区別し、構造を無理に一致させない**: 同じ「書籍」でも
  - **Domain Model**（業務ルールを表現。例: `domain/book.Book` + `Price` VO）
  - **Read Model**（参照専用に最適化。例: `domain/book.BookSummary`。命名は `XxxView` / `XxxReadModel` / `XxxListItem` / `XxxSummary` など用途が伝わるものでよい）
  - **Persistence Model**（DB 永続化専用。例: `infrastructure/postgres/book.model`、`gorm` タグ付き）
  はそれぞれ別目的であり、DB テーブル構造をそのまま Domain Model として扱ってはならない

変更容易性 = 影響範囲が局所化され、関係ない機能が壊れにくい状態

### CQRS（Repository と Query）

| | 書き込み側 | 読み取り側 |
|---|---|---|
| Domain IF | `<domain>.Repository` | `<domain>.Query` |
| UseCase | `usecase/<domain>/command` | `usecase/<domain>/query` |
| Infrastructure | `postgres/<domain>.NewRepository` | `postgres/<domain>.NewQuery` |

- 同一テーブルを両方から触ってよい（実装を1 struct にまとめて両方の IF を満たしてもよい）
- DB は1つ（アプリでの Writer/Reader 必須分離はしない）
- イベント／投影は必要なドメインだけ。分散イベントバス・全体 ES はしない
- 複数テーブルにまたがる更新など厳密整合は同一トランザクションの Command
- **更新系**: 境界は **Command DTO**（`usecase/<domain>/command` が定義。例: `CreateCommand` / `UpdateCommand`）で受け渡す。UseCase 内で DTO → ドメインモデル（Entity/Aggregate）へ変換してから Repository に渡す
  - Presentation の Request 構造体・ORM Entity・DB Entity・Framework 依存オブジェクト（`echo.Context` 等）を UseCase／Domain へ直接渡さない。Handler は Request から値を取り出して Command DTO を組み立てる
  - Command DTO 自身に業務ロジック（バリデーション・状態遷移・DB／外部 API アクセス）を書かない。責務はデータの受け渡しのみで、判定は Domain（VO のコンストラクタ等）に置く
- **参照系**: **Read Model**（参照専用モデル）中心で設計する。表示・検索用途に最適化した形を返し、更新系 DTO やドメインモデルを無理に再利用しない
- 更新モデルと参照モデルを混在させない（同じ struct を両方の用途に使い回さない）

### Query / ExternalGateway の IF 置き場

Repository／Query／ExternalGateway の IF はすべて **Domain** に置き、実装だけを Infrastructure に置く（UseCase には置かない）。種類によって置き場を変えると「なぜ IF によって置き場が違うのか」を調べ直すコストが生じるため、外部リソースへの抽象 IF は常に Domain に置くことで一貫させる。

### Command DTO を使う範囲

作成・更新するデータを受け取る UseCase は、フィールドが1つでも Command DTO を使う。ID のみを受け取る操作（Delete 等）は scalar 引数のままでよい（1 フィールドの入れ物は何の契約も表現しないため）。

## 3. HTTP / OpenAPI

**BE の正**: Go のリクエスト／レスポンス DTO（`validate` タグ）と薄い Handler（swag コメント付き）

| 要素 | 内容 |
|---|---|
| 入力 | 共通 `BindValidate[T]`（`Bind` + `validate.Struct`）。失敗時は 400 + フィールド単位 `errors[{field, rule}]`（`field` は JSON 名） |
| Handler | path パース → BindValidate（必要時）→ **UseCase** → JSON。エラーは `return err` するだけ |
| エラー | Echo の `HTTPErrorHandler` に1か所で登録（domain error → 4xx/5xx）。各 Handler に分岐を散らさない。5xx は error ログ＋メッセージ秘匿、4xx は warn ログ |
| OpenAPI | **swag で排出**（`backend/api/docs/` 等。手編集しない） |
| FE | 排出 OpenAPI → **openapi-typescript または orval**（必須） |
| 共通エラー形 | `ErrorResponse` を DTO で1か所 |
| ログ | `slog` のテキスト形式（デプロイしないので環境ごとに切り替えない）。アクセスログは RequestID 付き |
| 終了 | SIGINT / SIGTERM で graceful shutdown |

```text
Go DTO + validate + 薄い Handler + swag
  → swag init（OpenAPI 排出）
  → TypeScript 生成（必須）
  → FE は生成物のみ
  → CI: 排出 + TS gen → git diff --exit-code
```

### やらないこと（HTTP）

- ルート手書き `openapi/` を正とし **oapi-codegen で Server IF を生成**する方式（OpenAPI Design First）。契約を先に合意できる利点はあるが、全 Handler の書き換えになるため**保留**（[ADR](../docs/adr/2026-11-02-revise-development-rules.md)）。`internal/<業務領域>/<層>` への配置換えも同じ理由で保留
- **Huma**
- 各 Handler に Bind / validate / ステータス分岐を冗長コピペする方式
- FE で API 型を手書きして正とする運用
- TS／swag 排出物の手編集

### 漏れにくくする

- 文字数など業務制約は **domain VO** でも保証（HTTP タグだけに頼らない）
- HTTP ステータスは `HTTPErrorHandler` に集約

## 4. ディレクトリ（目安）

```
backend/
  cmd/
    api/main.go
  config/
  api/docs/                 # swag 排出 OpenAPI（手編集しない）
  migrations/               # golang-migrate
  internal/
    presentation/http/
      common/               # Echo・BindValidate・HTTPErrorHandler
      <domain>/             # Handler・dto
    usecase/
      <domain>/{command,query}/
    domain/
      common/               # sentinel error
      <domain>/             # Entity・VO・Repository/Query/ExternalGateway IF
    infrastructure/postgres/
      common/               # DB 接続
      <domain>/             # Repository/Query 実装
    infrastructure/gateway/  # 外部 API・SaaS 連携が必要になったら追加（ExternalGateway 実装）
  docker-compose.yml
  Makefile
  go.mod
  README.md
```

ドメインの package 分割は、境界が明確になってから（空の量産はしない）。

### プロダクト API 範囲

| Method | Path | 用途 | 認証 |
|---|---|---|---|
| `GET` | `/api/books` | 読んだ本の一覧（新しく登録した順。`limit` 1〜100・既定20、`offset`。`q`（書名・著者の部分一致）と `tagId`（分野タグ）で絞り込める。総件数は条件に合う件数。感想の本文は含めない） | 不要 |
| `GET` | `/api/books/{id}` | 読んだ本の詳細（書誌・書影・Amazon リンク・感想・評価） | 不要 |
| `POST` | `/api/books` | 読んだ本を登録する（`isbn`・`summary`・`comment`・`rating`、任意で `titleOverride`・分野タグ（`tagIds`）。書誌と書影は ISBN で外部カタログから取得。openBD に無ければ 400、同じ ISBN は 409） | 必要 |
| `PUT` | `/api/books/{id}` | 一言まとめ・感想・評価・書名の上書き・分野タグ（`tagIds`）を更新する（楽観的ロック。書誌と書影を取り直す） | 必要 |
| `DELETE` | `/api/books/{id}` | 読んだ本を削除する | 必要 |
| `GET` | `/api/catalog/{isbn}` | 登録前に、ISBN で外部カタログの書誌と書影を確かめる | 必要 |
| `GET` | `/api/tags` | 分野タグの一覧 | 不要 |
| `GET` | `/api/tags/counts` | 分野タグごとの冊数（本が付いていないタグは含めない。冊数の多い順、同数ならタグ名順） | 不要 |
| `POST` | `/api/tags` | 分野タグを追加する（同名は409） | 必要 |
| `PUT` | `/api/tags/{id}` | 分野タグの名前を変更する（楽観的ロック） | 必要 |
| `DELETE` | `/api/tags/{id}` | 分野タグを削除する（付いていた本からは自動で外れる） | 必要 |
| `POST` | `/api/auth/login` | ID とパスワードでログインし、セッション Cookie を発行する（失敗が続くと一時的に拒否） | 不要 |
| `POST` | `/api/auth/logout` | セッションを破棄して Cookie を消す | 不要 |
| `GET` | `/api/auth/session` | ログイン済みか確かめる（未ログインは 401） | 必要 |

- 認証は httpOnly Cookie `admin_session`。ID とパスワードのハッシュは環境変数（`ADMIN_ID` / `ADMIN_PASSWORD_HASH`）。自分だけが書き込めればよいので、ユーザー管理は持たない
- 書影は提供元の URL をそのまま返す（[ADR](../docs/adr/2026-09-28-book-cover-from-external-catalogs.md)）。登録と更新のたびに取り直す

## 5. やらないこと（全体）

- 全体 Event Sourcing / イベントストア正
- Command 用 DB と Query 用 DB の分離
- 分散イベントバス初期前提
- マルチサービス分割
- Dev Container

## 6. コマンド

日常の品質チェックは `make fmt-check lint vuln test`（CI と同じ。一覧は `make help`）。

### OpenAPI 排出

```bash
make tools          # swag / migrate / golangci-lint / govulncheck を版固定で導入
make swagger        # 排出
make swagger-check  # ドリフト検査（CI と同じ）
```

排出物は `api/docs/swagger.yaml` / `swagger.json`（手編集しない。`.gitattributes` の `linguist-generated` で PR の差分では折りたたまれる）。FE はこれを入力に TypeScript を生成する。
