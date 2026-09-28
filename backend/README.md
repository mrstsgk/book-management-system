# Backend

Go / Echo モノリス。方針の正は [`architecture.md`](./architecture.md)。

## ローカル開発

| 対象 | 方法 |
|---|---|
| Go | **ホスト**（版はリポジトリ直下 `.mise.toml` で固定） |
| DB | **Docker Compose**（PostgreSQL） |
| 画像ストレージ | **Docker Compose**（LocalStack の S3。バケット `book-images` は起動時に自動作成。ポートは `127.0.0.1` のみに公開するため Docker Engine 28.0.0 以上が必要） |
| 見本データ | `make seed` で書籍・著者（ID 9001 番台）を投入。表紙画像（`localstack/seed/`）は LocalStack の起動のたびに自動でアップロードされる |
| Dev Container | 使わない |

```bash
# リポジトリルート
mise install

cd backend
make tools          # swag / migrate / golangci-lint / govulncheck を版固定で導入
make db-up
make migrate-up
make seed           # 見本の書籍・著者（任意。何度実行しても同じ状態になる）
make run
# http://localhost:8080/health
# http://localhost:8080/api/authors/1/books

# OpenAPI 再排出（DTO/Handler 変更後）
make swagger
```

## 現状

- 書籍・著者 API（旧 Kotlin 実装相当。一覧は `architecture.md` §4）
- DB スキーマ: `migrations/`（資料は `docs/db/backend-schema.{json,md}`）
- Repository／Query／ImageStorage の契約テストはローカルの PostgreSQL・LocalStack（`make db-up migrate-up`）に対して実行し、起動していなければ skip する
- 書籍の表紙画像は `POST /api/books/{id}/image`（multipart の `image`）で受け取り S3 に保存する。DB にはオブジェクトキーだけを持ち、書籍取得時に15分有効の署名付き URL（`imageUrl`）を発行する
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 画像ストレージ | S3（aws-sdk-go-v2。ローカルは LocalStack 4.9） |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。

### LocalStack の画像について

LocalStack（コミュニティ版）はデータを永続化しないため、コンテナを作り直すとアップロードした画像は消える（DB には画像キーが残るので、その書籍の `imageUrl` は 404 になる）。見本データの表紙画像だけは起動のたびに自動でアップロードされるので、見本データで画面を確認する分には影響しない。自分でアップロードした画像が必要なら、作り直した後にアップロードし直す。
