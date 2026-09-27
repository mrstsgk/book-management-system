# Backend

Go / Echo モノリス。方針の正は [`architecture.md`](./architecture.md)。

## ローカル開発

| 対象 | 方法 |
|---|---|
| Go | **ホスト**（版はリポジトリ直下 `.mise.toml` で固定） |
| DB | **Docker Compose**（PostgreSQL） |
| Dev Container | 使わない |

```bash
# リポジトリルート
mise install

cd backend
make tools          # swag / migrate / golangci-lint / govulncheck を版固定で導入
make db-up
make migrate-up
make run
# http://localhost:8080/health
# http://localhost:8080/api/authors/1/books

# OpenAPI 再排出（DTO/Handler 変更後）
make swagger
```

## 現状

- 書籍・著者 API（旧 Kotlin 実装相当。一覧は `architecture.md` §4）
- DB スキーマ: `migrations/`（資料は `docs/db/backend-schema.{json,md}`）
- Repository／Query の契約テストはローカル DB（`make db-up migrate-up`）に対して実行し、DB が無ければ skip する
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。
