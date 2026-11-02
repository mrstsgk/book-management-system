# コードのコメントの書き方

実装にコメントを残すときの規則。生成物（`backend/api/docs/`・`frontend/*/src/api/generated/`）は対象外（手編集しないため）。本リポジトリのコードのコメントはすべて日本語で書く。

## 何をどこに書くか

同じ情報を複数の場所に書かないよう、書く場所ごとに役割を分ける。

| 置き場 | 書くこと |
|---|---|
| ソースコード | How（仕組み・ロジック） |
| テストコード | What（期待する振る舞い・仕様。[テスト規約](./testing.md)） |
| コミットログ | Why（変更の動機。[コミットメッセージ規約](./git-commit-message.md)） |
| **コードのコメント** | **Why not**（採らなかった案、コードから読み取れない制約）。ただし下の「概要のドキュメントコメント」は What |

## Why not

コードを読めば分かる処理の説明は書かない。書くのは、読んだ人が「なぜこう書かなかったのか」と迷う箇所だけにする。

- 採らなかった案と、採らなかった理由
- コードから読み取れない制約（外部 API の仕様、ライブラリの不具合の回避、版を固定している理由など）

```go
// 書影の URL は保存しない（openBD 側の削除・差し替えを次回の取り直しで反映するため）
cover := catalog.CoverURL
```

## 概要のドキュメントコメント

下の「バックエンド」「フロントエンド」の表にあるものには、何の知識か・何をするかの概要を、日本語のドキュメントコメントで 1 行書く。読む人が、ほかのファイルを開かずに、その場やエディタのホバーで意味を確かめられるようにするため。

- 概要だけを書き、制約・条件（文字数、重複しないこと、件数の上限など）は書かない。詳細はコード（VO のコンストラクタ・UseCase・`validate` タグなど）で表す。コメントに書くと、コードを直したときにコメントだけが古くなるため
- 同じ層の中では書き方を揃える。種類ごとに書く・書かないを変えると、どれに当たるかを見分ける手間が増えるため
- 層を問わず、複雑な関数・メソッド（流れが追いにくい、分岐が多い、自明でないアルゴリズム）にも 1 行の概要を書く

ドキュメントコメントの形は言語ごとの標準に従う（`go doc`・エディタのホバーに表示されるため）。

| 言語 | 形 |
|---|---|
| Go | 宣言の直前の行に `//`（空行を挟まない）。型・関数・メソッドは識別子の名前で始める（`// Book は、…`）。フィールドは概要だけでよい |
| TypeScript | 宣言の直前に TSDoc の `/** … */`。値の集合（文字列のユニオン型）の各値の意味は、型のコメントに箇条書きで書く（ユニオンの各値に付けたコメントはホバーに出ないため） |

以下の例の名前と型は説明のためのもので、実装の決定ではない。UseCase の例は [`backend/architecture.md`](../../backend/architecture.md) の命名規則（`<動詞><業務対象>Usecase`）で書いている。

## バックエンド

層・ディレクトリは `internal/{presentation/http,usecase,domain,infrastructure/postgres,infrastructure/gateway}/<domain>`（[`backend/architecture.md`](../../backend/architecture.md) §2）。

| 層 | 1 行の概要を書くもの |
|---|---|
| Domain（Entity・VO・Read Model） | 型と各フィールド |
| Domain（Repository・Query・ExternalGateway の IF） | 型と各メソッド |
| UseCase | IF の型（メソッドは `Execute` だけなので書かない）。Command DTO の型と各フィールド |
| Presentation（Handler） | 型と各メソッド。メソッドには HTTP メソッドとパスを添える。Request／Response DTO（`dto.go`）には書かない（swag の説明タグと `validate` タグで表す） |
| Infrastructure | Repository・Query・ExternalGateway の実装の型と各メソッド |
| 上記以外（`cmd`・`presentation/http/common`・`infrastructure/postgres/common`・変換関数など） | 書かない（Why not だけ） |

UseCase の実装の型（`XxxUsecaseImpl`）は IF を見れば分かるので書かない。Persistence Model（`gorm` タグの付いた型）はカラムの意味の正が migration の `COMMENT ON COLUMN` と [`docs/db/backend-schema.md`](../db/backend-schema.md)（[DB 資料の規則](./db-documentation.md)）なので書かない。

### Domain（Entity・VO・Read Model）

```go
// Book は、自分が読んだ本と、それに添える感想・評価。
type Book struct {
	// 本を区別する ISBN。
	ISBN ISBN
	// 画面に出す書名（上書きがあればそれ）。
	Title Title
	// 一言まとめ・感想・評価。
	Review Review
	// 同時編集を見つけるための版。
	Version int64
}

// Title は、本の書名。
type Title struct{ value string }

// BookSummary は、一覧の 1 行。
type BookSummary struct {
	// 本を区別する ID。
	ID int64
	// 画面に出す書名。
	Title string
}
```

### Domain（Repository・Query・ExternalGateway の IF）

```go
// Repository は、読んだ本を保存する。
type Repository interface {
	// FindByID は、ID で本を 1 件取り出す。
	FindByID(ctx context.Context, id int64) (*Book, error)
}

// Query は、本の一覧と詳細を画面に出すために読み取る。
type Query interface {
	// Search は、条件に合う本を新しい順に探す。
	Search(ctx context.Context, cond ListCondition) ([]BookSummary, error)
}

// BookCatalog は、ISBN から書誌と書影を取ってくる外部カタログ。
type BookCatalog interface {
	// Lookup は、ISBN で書誌と書影を 1 件探す。
	Lookup(ctx context.Context, isbn ISBN) (*Bibliography, error)
}
```

### UseCase

```go
// RegisterBookUsecase は、読んだ本を新しく登録する。
type RegisterBookUsecase interface {
	Execute(ctx context.Context, cmd RegisterCommand) (int64, error)
}

type RegisterBookUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
}

// RegisterCommand は、本の新規登録の入力。
type RegisterCommand struct {
	// 登録する本の ISBN。
	ISBN string
	// 一覧で見せる一言まとめ。
	Summary string
	// 付ける分野タグの ID。
	TagIDs []int64
}
```

### Presentation（Handler）

```go
// Handler は、読んだ本の API を処理する。
type Handler struct {
	SearchBooksUsecase  bookquery.SearchBooksUsecase
	RegisterBookUsecase bookcmd.RegisterBookUsecase
}

// List は、読んだ本の一覧を返す（GET /api/books）。
func (h *Handler) List(c echo.Context) error { ... }

// Register は、読んだ本を登録する（POST /api/books）。
func (h *Handler) Register(c echo.Context) error { ... }
```

### Infrastructure

```go
// repository は、読んだ本を PostgreSQL に保存する。
type repository struct{ db *gorm.DB }

// FindByID は、ID で本を 1 件取り出す。
func (r *repository) FindByID(ctx context.Context, id int64) (*book.Book, error) { ... }

// query は、本の一覧と詳細を PostgreSQL から読み取る。
type query struct{ db *gorm.DB }

// Search は、条件に合う本を新しい順に探す。
func (q *query) Search(ctx context.Context, cond book.ListCondition) ([]book.BookSummary, error) { ... }
```

## フロントエンド

| 対象 | 1 行の概要を書くもの |
|---|---|
| 型（UI Model・Props・値の集合） | 型と各フィールド。値の集合は各値の意味 |
| Component | 関数（何を表示し、何をさせるか） |
| hook（feature の `hooks/`・store） | 関数（何を提供するか） |
| utils の関数 | 関数（何をするか） |
| 上記以外（`app/`・`lib/`・`testing/`） | 書かない（Why not だけ） |

```ts
/**
 * 一覧の並び順。
 * - `newest`: 新しく登録した順
 * - `rating`: 評価の高い順
 */
export type SortKey = 'newest' | 'rating'

/** 一覧の 1 冊分の表示項目。 */
export interface BookListItem {
  /** 画面に出す書名。 */
  title: string
  /** 書影の URL。無ければ書影なしの枠を出す。 */
  coverUrl?: string
}

/** 読んだ本を一覧に並べ、検索とタグの絞り込みを受け付ける。 */
export function BooksListView({ items, onSearch }: BooksListViewProps) { ... }

/** 本の一覧を取得し、検索・絞り込み・追加読み込みの状態を提供する。 */
export function useBooksList() { ... }

/** タグ名から選択済みのタグ ID を引く。 */
export function tagIdsByName(tags: Tag[], names: string[]): number[] { ... }
```
