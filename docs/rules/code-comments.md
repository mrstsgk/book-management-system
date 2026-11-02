# コードのコメントの書き方

実装にコメントを残すときの規則。OpenAPI などから自動生成したファイルは対象外（手編集しないため）。

## 何をどこに書くか

同じ情報を複数の場所に書かないよう、書く場所ごとに役割を分ける。

| 置き場 | 書くこと |
|---|---|
| ソースコード | How（仕組み・ロジック） |
| テストコード | What（期待する振る舞い・仕様。[テスト規約](./testing.md)） |
| コミットログ | Why（変更の動機。[コミットメッセージ規約](./git-commit-message.md)） |
| **コードのコメント** | **Why not**（採らなかった案、コードから読み取れない制約）。ただし下の「概要のドキュメントコメント」は What |

コメントはすべて日本語で書く。

## Why not

コードを読めば分かる処理の説明は書かない。書くのは、読んだ人が「なぜこう書かなかったのか」と迷う箇所だけにする。

- 採らなかった案と、採らなかった理由
- コードから読み取れない制約（外部 API の仕様、ライブラリの不具合の回避、版を固定している理由など）

```go
// 生成 IF を埋め込まないのは、未実装の operation が実行時エラーになるため
var _ adminapi.ServerInterface = (*server)(nil)
```

## 概要のドキュメントコメント

下の「バックエンド」「フロントエンド」の表にあるものには、何の知識か・何をするかの概要を、日本語のドキュメントコメントで 1 行書く。読む人が、ほかのファイルを開かずに、その場やエディタのホバーで意味を確かめられるようにするため。

- 概要だけを書き、制約・条件（文字数、重複しないこと、件数の上限など）は書かない。詳細はコード（VO のコンストラクタ・UseCase・zod など）で表す。コメントに書くと、コードを直したときにコメントだけが古くなるため
- 同じ層の中では書き方を揃える。種類ごとに書く・書かないを変えると、どれに当たるかを見分ける手間が増えるため
- 層を問わず、複雑な関数・メソッド（流れが追いにくい、分岐が多い、自明でないアルゴリズム）にも 1 行の概要を書く

ドキュメントコメントの形は言語ごとの標準に従う（`go doc`・エディタのホバーに表示されるため）。

| 言語 | 形 |
|---|---|
| Go | 宣言の直前の行に `//`（空行を挟まない）。型・関数・メソッドは識別子の名前で始める（`// Book は、…`）。フィールドは概要だけでよい |
| TypeScript | 宣言の直前に TSDoc の `/** … */`。値の集合（文字列のユニオン型）の各値の意味は、型のコメントに箇条書きで書く（ユニオンの各値に付けたコメントはホバーに出ないため） |

以下の例の名前と型は説明のためのもので、実装の決定ではない。

## バックエンド

| 層 | 1 行の概要を書くもの |
|---|---|
| Domain（Entity・VO・Read Model） | 型と各フィールド |
| Domain（Repository・Query・ExternalGateway の IF） | 型と各メソッド |
| UseCase | IF の型（メソッドは `Execute` だけなので書かない）。Command DTO の型と各フィールド |
| Presentation（Handler） | 型と各メソッド。メソッドには HTTP メソッドとパスを添える |
| Infrastructure | Repository・Query・ExternalGateway の実装の型と各メソッド |
| 上記以外（`cmd`・`platform`・`domainerr`・変換関数など） | 書かない（Why not だけ） |

UseCase の実装の型（`XxxUsecaseImpl`）は IF を見れば分かるので書かない。Persistence Model（`gorm` タグの付いた型）はカラムの意味の正が migration の `COMMENT ON COLUMN` と [`docs/db/backend-schema.md`](../db/backend-schema.md)（[DB 資料の規則](./db-documentation.md)）なので書かない。

### Domain（Entity・VO・Read Model）

```go
// Book は、読んだ本と、それに添えた感想・評価・分野タグ。
type Book struct {
	// 本を区別する ID。
	ID ID
	// 本を特定する ISBN。
	ISBN ISBN
	// 一覧に出す一言まとめ。
	Summary Summary
	// 同時更新を見つけるための版。
	Version int
}

// Summary は、一覧に出す一言まとめ。
type Summary struct {
	// まとめの本文。
	value string
}

// BookListItem は、本の一覧の 1 行。
type BookListItem struct {
	// 本を区別する ID。
	ID int64
	// 画面に出す書名。
	Title string
}
```

### Domain（Repository・Query の IF）

```go
// Repository は、読んだ本を保存する。
type Repository interface {
	// FindByID は、ID で本を 1 件取り出す。
	FindByID(ctx context.Context, id ID) (*Book, error)
}

// Query は、本の一覧と詳細を画面に出すために読み取る。
type Query interface {
	// Search は、条件に合う本を探す。
	Search(ctx context.Context, c ListCondition, r ListRange) (*BookList, error)
}
```

### UseCase

```go
// RegisterBookUsecase は、読んだ本を新しく登録する。
type RegisterBookUsecase interface {
	Execute(ctx context.Context, cmd RegisterBookCommand) (*domain.BookDetail, error)
}

type RegisterBookUsecaseImpl struct {
	Books domain.Repository
}

// RegisterBookCommand は、読んだ本の登録の入力。
type RegisterBookCommand struct {
	// 登録する本の ISBN。
	ISBN string
	// 一覧に出す一言まとめ。
	Summary string
	// 付ける分野タグの ID。
	TagIDs []int64
}
```

### Presentation（Handler）

```go
// Handler は、管理画面の本の API を処理する。
type Handler struct {
	SearchBooksUsecase  bookqry.SearchBooksUsecase
	RegisterBookUsecase bookcmd.RegisterBookUsecase
}

// SearchBooks は、本を検索する（GET /api/admin/books）。
func (h *Handler) SearchBooks(c echo.Context, params adminapi.SearchBooksParams) error { ... }

// RegisterBook は、読んだ本を新しく登録する（POST /api/admin/books）。
func (h *Handler) RegisterBook(c echo.Context) error { ... }
```

### Infrastructure

```go
// repository は、読んだ本を PostgreSQL に保存する。
type repository struct{ db *gorm.DB }

// FindByID は、ID で本を 1 件取り出す。
func (r *repository) FindByID(ctx context.Context, id domain.ID) (*domain.Book, error) { ... }

// query は、本の一覧と詳細を PostgreSQL から読み取る。
type query struct{ db *gorm.DB }

// Search は、条件に合う本を探す。
func (q *query) Search(ctx context.Context, c domain.ListCondition, r domain.ListRange) (*domain.BookList, error) { ... }
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
export type BookSort = 'newest' | 'rating'

/** 本の一覧の 1 枚のカード。 */
export interface BookCardProps {
  /** 画面に出す書名。 */
  title: string
  /** 押したときの移動先。 */
  to: string
}

/** 本の書影・書名・一言まとめを 1 枚のカードで見せる。 */
export function BookCard({ title, to }: BookCardProps) { ... }

/** 本の一覧を取得し、画面用の形にして提供する。 */
export function useBooksList() { ... }

/** 文字数を Unicode のコードポイントで数える。 */
export function codePointLength(value: string): number { ... }
```
