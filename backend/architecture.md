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
| 書誌・書影 | 外部カタログ。Domain の `book.BookCatalog`（ExternalGateway）を `infrastructure/gateway/openbd`・`rakuten` が実装し、`gateway/catalog` が「書誌は openBD だけから取り、書影が無ければ楽天で補う」形に組み合わせる。画像は保存しない |
| 認証 | 書き込み系だけ管理者トークン（`presentation/http/common.RequireAdminToken`）。閲覧は認証なし |
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
- 命名目安: `CreateUsecase` / `CreateUsecaseImpl`、`Handler`、`dto.go`
- `main` で手書き DI（コンテナなし）: config → infra → usecase → presentation
- 外部 API・SaaS 呼び出しは UseCase から **ExternalGateway IF**（Domain 定義）経由で行う。Repository／Query と同じ理由（Domain が具体的な HTTP クライアント実装に依存しないため）で Domain に IF を置く

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

- ルート手書き `openapi/` を正とし **oapi-codegen で Server IF を生成**する方式
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

- 認証は `Authorization: Bearer <ADMIN_TOKEN>`。自分だけが書き込めればよいので、ユーザー管理は持たない
- 書影は提供元の URL をそのまま返す。`coverSource` が `rakuten` なら画面に楽天ウェブサービスのクレジット表示が必要（[ADR](../docs/adr/2026-09-28-book-cover-from-external-catalogs.md)）
- 楽天の書影は楽天の商品ページ（`coverProductUrl`）と一緒に返す。取得から89日以上経過した楽天の書影は返さない（楽天の規約の保持期限）（[設計](../docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md)）
- 楽天の書影は、API の起動時・稼働中は1日1回・本の更新時に、期限が近いものを取り直す。取り直せないまま期限を過ぎたら外す

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

排出物は `api/docs/swagger.yaml` / `swagger.json`（手編集しない）。FE はこれを入力に TypeScript を生成する。
