# バックエンド DB スキーマ

正は `backend/migrations/*.sql`。AI 向けの詳細（型・nullable・値の意味）は [`backend-schema.json`](./backend-schema.json)。ルールは [`docs/rules/db-documentation.md`](../rules/db-documentation.md)。

`book`: 自分が読んだ本。書誌（書名・著者・出版社・発売日）と書影の URL は ISBN で外部カタログから取得した値、感想と評価は自分で書いた値。著者は提供元の文字列のまま持つ（独立したテーブルは無い）。ISBN は一意。楽天の書影は商品ページの URL と取得日時（保持期限89日の起点）を必ず一緒に持つ。

```mermaid
erDiagram
    book {
        bigserial id PK
        varchar_13 isbn "unique"
        varchar_255 title
        varchar_255 title_override "nullable"
        varchar_500 authors
        varchar_255 publisher
        varchar_32 published_on
        varchar_2048 cover_url "nullable"
        varchar_16 cover_source "nullable"
        varchar_2048 cover_product_url "nullable"
        timestamptz cover_fetched_at "nullable"
        text comment
        smallint rating
        varchar_100 summary
        integer version
        timestamptz created_at
        timestamptz updated_at
    }
```

`tag`: 自分が定義した分野タグ。本には `book_tag` 経由で複数付けられる。タグ名は一意。

`book_tag`: 本と分野タグの多対多の中間テーブル。本・タグどちらが消えても自動で外れる（`ON DELETE CASCADE`）。

```mermaid
erDiagram
    tag {
        bigserial id PK
        varchar_30 name "unique"
        integer version
        timestamptz created_at
        timestamptz updated_at
    }
    book_tag {
        bigint book_id PK
        bigint tag_id PK
    }
    book ||--o{ book_tag : "付く"
    tag ||--o{ book_tag : "付く"
```
