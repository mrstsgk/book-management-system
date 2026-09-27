# バックエンド DB スキーマ

正は `backend/migrations/*.sql`。AI 向けの詳細（型・nullable・値の意味）は [`backend-schema.json`](./backend-schema.json)。ルールは [`docs/rules/db-documentation.md`](../rules/db-documentation.md)。

- `book`: 書籍情報。価格は円の整数、出版状況は 1: 未出版 / 2: 出版済み。Amazon の商品リンクと表紙画像のストレージ上のキーは任意。
- `author`: 著者情報。生年月日は任意。
- `author_book`: 著者と書籍の多対多の関連。書籍には著者が1人以上紐づく（アプリ側で保証）。

```mermaid
erDiagram
    book {
        serial id PK
        varchar_255 title
        numeric_10_2 price
        integer publish_status
        varchar_2048 amazon_url "nullable"
        varchar_255 image_key "nullable"
        integer version
    }
    author {
        serial id PK
        varchar_100 name
        date birth_date "nullable"
        integer version
    }
    author_book {
        integer author_id PK, FK
        integer book_id PK, FK
        integer version
    }
    author ||--o{ author_book : "著す"
    book ||--|{ author_book : "著者を持つ"
```
