# 一覧の検索・絞り込みと分野別の集計 設計

**日付:** 2026-09-28
**対象:** バックエンド拡張（[要求定義](../../requirements.md) §5 単位1 の「分野別の集計」「キーワード検索」）。画面側の要件は[要件定義](../../specifications.md) §2.1・§2.3
**変更前の形:** [読んだ本の API 設計](./2026-09-28-reading-api-design.md)、[分野タグ 設計](./2026-09-28-book-tags-design.md)。この文書はそこからの差分だけを書く

## 目的

- 公開画面の一覧で、キーワード（書名・著者）の検索と分野タグの絞り込みを同時に使えるようにする
- 分野タグごとの冊数を返し、そこから選んだ分野で絞り込んだ一覧へ移れるようにする

## やらないこと

- 複数タグでの絞り込み（AND/OR）。要件にない
- 全文検索・形態素解析・表記ゆれの吸収。個人の読書記録の件数では部分一致で足りる
- カーソル方式のページング。既存の `limit`/`offset` で「続きを読み込む」「条件を変えたら先頭から」を満たせる
- 画面そのもの（単位2）

## 1. 一覧の検索・絞り込み（`GET /api/books` の拡張）

新しいエンドポイントは作らず、既存の一覧にクエリパラメータを足す。検索・絞り込みは一覧そのものの条件であり、分けると画面が2つの API を使い分けることになるため。

| パラメータ | 型 | 省略時 | 意味 |
|---|---|---|---|
| `q` | string | 絞り込まない | 書名・著者の部分一致（大文字小文字を区別しない） |
| `tagId` | int | 絞り込まない | その分野タグが付いた本だけ |
| `limit` / `offset` | int | 既存どおり | 変更なし |

- `q` と `tagId` は同時に使える（AND）
- `total` は条件に合う件数（取得範囲外も含む）。条件を変えたら画面は `offset=0` から取り直す
- 書名は表示と同じく、上書きがあれば上書きを検索対象にする（`COALESCE(title_override, title)`）。上書き前の書名では当たらない（画面に見えていない書名で当たると利用者が混乱するため）
- `q` に含まれる `%`・`_`・`\` は文字として扱う（LIKE のワイルドカードとして解釈しない）
- 存在しない `tagId` はエラーにせず0件を返す（絞り込みの条件であり、作ろうとしている資源ではないため）

### 1.1 ドメイン（参照側の条件）

`domain/book` に一覧の条件を表す VO を足す。`common.ListRange` と同じく参照専用の入力で、書き込み側の VO（`Title` など）は使わない。

```go
// ListCondition は一覧の検索・絞り込みの条件。ゼロ値は「絞り込まない」。
type ListCondition struct {
	keyword string  // 前後の空白を除いたもの。空なら検索しない
	tagID   *tag.ID // nil なら絞り込まない
}

func NewListCondition(keyword string, tagID *tag.ID) (ListCondition, error) // keyword は100文字まで、制御文字は不可
```

`Query.FindList` のシグネチャを `FindList(ctx, c ListCondition, r common.ListRange) (*BookList, error)` に変える。

### 1.2 ユースケース

`ListUsecase.Execute(ctx, limit, offset int)` を、入力の構造体を受け取る形に変える（引数が4つになり、同じ型の `int` が並ぶと取り違えやすいため）。

```go
type ListInput struct {
	Keyword string
	TagID   *int64
	Limit   int
	Offset  int
}
```

参照系の入力に構造体を使うのは本リポジトリで初めて。Command DTO と同じく「データの受け渡しだけ・判定は VO」とする。

### 1.3 永続化

- `q`: `(COALESCE(title_override, title) ILIKE ? ESCAPE '\' OR authors ILIKE ? ESCAPE '\')`
- `tagId`: `EXISTS (SELECT 1 FROM book_tag WHERE book_tag.book_id = book.id AND book_tag.tag_id = ?)`（JOIN にすると行が重複しうるので EXISTS）
- 総件数（`COUNT`）にも同じ条件を掛ける
- インデックスは足さない（件数が少なく、部分一致は通常の B-tree では効かない）

### 1.4 HTTP

`ListRequest` に `Q string \`query:"q" validate:"max=100"\`` と `TagID *int64 \`query:"tagId" validate:"omitempty,min=1"\`` を足す。

## 2. 分野別の集計（`GET /api/tags/counts`）

既存の `GET /api/tags`（管理画面のタグ一覧）とは目的が違うため、別のエンドポイントにする。管理画面の一覧に冊数を混ぜると、「なぜタグ一覧が冊数を返すのか」が後から分からなくなる。

```json
{ "items": [ { "id": 1, "name": "設計", "bookCount": 3 } ] }
```

- 本が1冊も付いていないタグは含めない（「分野別に何冊読んだか」の画面で0冊の分野を見せる意味がないため）
- 冊数の多い順、同数ならタグ名順
- 認証不要（公開画面で使う）
- 集計から分野を選んだときの一覧は、§1 の `GET /api/books?tagId=` をそのまま使う（この API は `id` を返すだけでよい）

### 2.1 ドメイン・ユースケース・永続化

| 層 | 追加 |
|---|---|
| `domain/tag` | Read Model `TagBookCount{ID ID; Name string; BookCount int}`・`TagBookCounts{Items []*TagBookCount}`、`Query` に `CountBooks(ctx) (*TagBookCounts, error)` |
| `usecase/tag/query` | `CountBooksUsecase`（`Query.CountBooks` をそのまま呼ぶ） |
| `postgres/tag` | `SELECT tag.id, tag.name, COUNT(*) FROM tag JOIN book_tag ON ... GROUP BY tag.id, tag.name ORDER BY COUNT(*) DESC, tag.name` の1クエリ |
| `presentation/http/tag` | `GET /counts` を認証なしで登録 |

集計の主語はタグなので `domain/tag` に置く（既存の `ExistsAll` と同じくタグの参照系）。

## 3. エラー

| 状況 | 結果 |
|---|---|
| `q` が101文字以上・制御文字を含む | 400 |
| `tagId` が0以下・数値でない | 400 |
| 存在しない `tagId` | 200・0件 |

## 4. テスト

| 対象 | 確かめること |
|---|---|
| `book.ListCondition` | 空・空白だけは検索しない、100文字ちょうど・101文字、制御文字、前後の空白を除く |
| `book.Query`（契約テスト） | 書名・著者の部分一致、大文字小文字の無視、上書きした書名で当たり上書き前の書名では当たらない、`%`・`_` を文字として扱う、タグ絞り込み、両方同時、`total` が条件に合う件数、タグが複数付いた本が重複しない |
| `tag.Query.CountBooks`（契約テスト） | 冊数、0冊のタグを含まない、並び順 |
| ユースケース（Fake） | 入力が不正ならクエリを呼ばない、条件がそのまま渡る |
| Handler | `q`・`tagId` の400、集計が認証なしで取れる |

## 5. PR の分け方

1. 一覧の検索・絞り込み（§1。`ListCondition`・`FindList` の変更・`ListInput`・Handler・OpenAPI/フロント再生成）
2. 分野別の集計（§2。1 と独立して進められる）

2本は触るファイルがほぼ重ならないので並行して進める。重なるのは OpenAPI とフロントの生成物だけで、後からマージする側で再生成すれば解消する。
