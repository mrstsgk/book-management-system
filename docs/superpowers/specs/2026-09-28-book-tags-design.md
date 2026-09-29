# 分野タグ 設計

**日付:** 2026-09-28
**対象:** バックエンド拡張 c（[要件定義](../../specifications.md) §1.3・§3.4、および §1.1「分野タグ」・§3.2・§3.3 の本への付与）
**変更前の形:** [読んだ本の API 設計（既存）](./2026-09-28-reading-api-design.md)、[書名の上書き・一言まとめ 設計](./2026-09-28-title-override-summary-design.md)。この文書はそこからの差分だけを書く

## 目的

- 自分が定義した分野タグを本に複数付けられるようにする（要求定義「学び続けている人」の根拠の1つ）
- タグの追加・名前の変更・削除ができる管理画面向けの API を用意する

## やらないこと

- 分野別の集計 API（単位 d）
- 公開画面・管理画面そのもの（単位 2・3）
- タグの並び順・色・階層化など、要件に無い管理機能

## 1. テーブル

PR の分け方（§8）に合わせ、マイグレーションを2つに分ける。

**`000003`（PR 2 で足す）:**

```sql
CREATE TABLE tag (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(30) NOT NULL,
    version    INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tag_name UNIQUE (name)
);
```

**`000004`（PR 3 で足す）:**

```sql
CREATE TABLE book_tag (
    book_id BIGINT NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, tag_id)
);
```

- `book_tag` は本・タグどちらが消えても自動で外れる（`ON DELETE CASCADE`）。アプリ側で明示的な削除処理を書かない
- `docs/db/backend-schema.{json,md}` に、それぞれのマイグレーションと同じ PR でテーブルを追加する

## 2. ドメイン

### 2.1 `domain/tag`（新しい集約）

`book` パッケージとは目的が違う（タグの命名規則は本の書名と揃える理由が無い）ため、検証は独立して持つ。

| 型 | 役割 |
|---|---|
| `ID` | `type ID int64` |
| `Name` | VO。前後の空白を除いて1〜30文字。制御文字は不可（`book.Title` と同じ形の検証だが、依存はしない） |
| `Tag` | エンティティ。`ID`・`Name`・`Version` |

```go
func New(name Name) *Tag                     // 登録前のタグを作る（ID・Version は保存時）
func (t *Tag) Rename(name Name, version int) // 名前を差し替える。version は楽観的ロック
```

`Repository`（書き込み）:

```go
type Repository interface {
	FindByID(ctx context.Context, id ID) (*Tag, error)     // ErrNotFound
	Create(ctx context.Context, t *Tag) error              // 同名は ErrConflict
	Update(ctx context.Context, t *Tag) error               // version 不一致・同名は ErrConflict
	Delete(ctx context.Context, id ID) error                // ErrNotFound
}
```

`Query`（参照）:

```go
type TagListItem struct { ID ID; Name string }
type TagList struct { Items []*TagListItem }

type Query interface {
	FindList(ctx context.Context) (*TagList, error)
	// ExistsAll は ids がすべて実在するかを返す（登録・更新時の入力検証に使う）
	ExistsAll(ctx context.Context, ids []ID) (bool, error)
}
```

一覧は取得件数を絞らない（管理画面専用で、想定件数は数十程度）。

### 2.2 `domain/book` の拡張

| 追加・変更 | 内容 |
|---|---|
| `TagSelection`（新しい VO） | `[]tag.ID` を包む。0〜10個、重複不可。0個は「タグ無し」として有効 |
| `Book.Tags` | `TagSelection` を追加 |
| `New` | `New(isbn, bibliography, cover, summary, comment, rating, tags TagSelection) *Book` に `tags` を足す |
| `ChangeReview` | `ChangeReview(summary, comment, rating, tags TagSelection, version int)` に `tags` を足す（タグも感想・評価と同じく PUT で一括更新するため。書名の上書きだけ `OverrideTitle` で別扱いのまま） |

`domain/book` は `domain/tag` の `tag.ID` 型だけを参照する（タグの実体は持たず、ID の集合だけを持つ）。存在確認（`ExistsAll`）はユースケースが担う。

```go
type TagSelection struct{ ids []tag.ID }

func NewTagSelection(ids []tag.ID) (TagSelection, error) // 上限10、重複はエラー
func (s TagSelection) IDs() []tag.ID
```

### 2.3 Read Model

| Read Model | 追加 |
|---|---|
| `BookDetail` | `Tags []string`（タグ名。無ければ空スライス） |
| `BookListItem` | `Tags []string` |

タグ名の取得は、対象の本の行を確定させた後に `book_tag` と `tag` を結合した別クエリで取り、Go 側でマージする（一覧のページングと1発の集計 JOIN が相性が悪いため。既存の「総件数を別クエリで数える」方針と同じ考え方）。

## 3. 永続化

### 3.1 `infrastructure/postgres/tag`（新規）

`postgres/book` と同じ形（`model`・`repository`・`query`・`adapt`）。`Create`/`Update` は一意制約違反（`uq_tag_name`）を `ErrConflict` に変換する。

### 3.2 `infrastructure/postgres/book` の拡張

- `Repository.Create`/`Update` はトランザクションにする（`db.Transaction(func(tx *gorm.DB) error {...})`）: 本の行を書いた後、`book_tag` を「Create は挿入のみ」「Update はいったん全削除して入れ直す」で同期する
- `Query.FindDetailByID`/`FindList` は、本の行を取った後に該当する `book_id` の `book_tag` × `tag` を `IN (...)` で1回取り、`map[book.ID][]string` に組んでから Read Model へ詰める

## 4. ユースケース

| ユースケース | 追加 |
|---|---|
| 登録 | `RegisterCommand.TagIDs []int64` を検証（重複・上限は `NewTagSelection`）→ `tag.Query.ExistsAll` で実在確認 → 無ければ `ErrInvalid`（「存在しないタグが含まれています」）。カタログ問い合わせより前に検証する（既存の方針どおり） |
| 更新 | `UpdateCommand.TagIDs []int64` も同様に検証・実在確認してから `ChangeReview` に渡す |
| タグの登録・名前変更・削除 | `tag.Repository` をそのまま呼ぶ薄いユースケース（`usecase/tag/command`） |
| タグ一覧 | `tag.Query.FindList` をそのまま呼ぶ（`usecase/tag/query`） |

## 5. HTTP / OpenAPI

### 5.1 `presentation/http/tag`（新規）

| Method | Path | 認証 |
|---|---|---|
| GET | `/api/tags` | 不要 |
| POST | `/api/tags` | 必要 |
| PUT | `/api/tags/{id}` | 必要 |
| DELETE | `/api/tags/{id}` | 必要 |

`TagRequest{Name string}`（`required,max=30`）、`UpdateTagRequest{Name string, Version *int}`、`TagResponse{ID, Name, Version}`、`TagListResponse{Items []TagResponse}`。

### 5.2 `presentation/http/book` の拡張

- `RegisterRequest`/`UpdateRequest` に `TagIDs []int64` を足す（`validate:"max=10"` で個数の上限だけ見る。要素が実在するかは `tag.Query.ExistsAll` に任せ、HTTP層で二重に検証しない。0個・省略は「タグ無し」）
- `Response`/`ListItemResponse` に `Tags []string` を足す
- `RegisterCommand`/`UpdateCommand` に `TagIDs []int64` を足す

## 6. エラー

| 状況 | 結果 |
|---|---|
| タグ名が重複 | 409 |
| タグの `version` 不一致 | 409 |
| 存在しないタグ ID | 404（`GET/PUT/DELETE /api/tags/{id}`）または 400（本の登録・更新に含めたとき） |
| タグを11個以上指定 | 400 |
| 同じタグ ID を重複指定 | 400 |

## 7. テスト

| 対象 | 確かめること |
|---|---|
| `tag.Name` | 上限ちょうど・上限+1、下限（空・空白だけ）、トリム、制御文字 |
| `book.TagSelection` | 0個・10個・11個、重複ID |
| `tag.Repository`（契約テスト） | 作成・改名・削除、同名の重複が409、version不一致が409、削除後に`book_tag`も消える（実際にJOINして確認） |
| `tag.Query`（契約テスト） | 一覧、`ExistsAll`（全部実在／一部欠け／空配列） |
| `book.Repository`（契約テスト） | タグ付きで作成・更新して読み込める、タグを空にすると外れる、タグを削除すると本から自動で外れる（`tag.Repository.Delete`経由） |
| `book.Query`（契約テスト） | 詳細・一覧がタグ名を返す。タグが無い本は空スライス（nilではない） |
| ユースケース（Fake） | 存在しないタグIDで登録・更新するとカタログ・保存を呼ばない。タグの登録・改名・削除の正常系とエラー変換 |
| Handler | `tagIds`の上限超え・重複が400、タグの一覧・登録・改名・削除の認証要否 |

## 8. PR の分け方

1. タグのドメイン（`domain/tag` の `Name`・`Tag`・`Repository`/`Query` の IF だけ）
2. タグの API 一式（マイグレーション `tag` テーブル・`infrastructure/postgres/tag`・`usecase/tag`・`presentation/http/tag`・OpenAPI・フロント再生成・DB資料）。本には触れない、独立して動く機能
3. 本とタグの結びつけ（`TagSelection`・`Book.Tags`・`book_tag`マイグレーション・`book.Repository`/`Query`のJOIN・登録更新のタグ検証・`Response`へのタグ名追加・OpenAPI/フロント再生成・DB資料）
