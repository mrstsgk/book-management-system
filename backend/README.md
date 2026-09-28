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
# http://localhost:8080/api/books
# 登録: curl -H "Authorization: Bearer local-admin-token" -H "Content-Type: application/json" \
#   -d '{"isbn":"9784873118703","comment":"感想","rating":5}' http://localhost:8080/api/books

# OpenAPI 再排出（DTO/Handler 変更後）
make swagger
```

## 現状

- 読んだ本の API（一覧・詳細・登録・更新・削除）とカタログの確認 API（範囲は `architecture.md` §4）
- 書誌・書影は ISBN で openBD から取得し、書影が無ければ楽天ブックスで補う（楽天は任意）
- 書き込み系は管理者トークンが必要（`Authorization: Bearer <ADMIN_TOKEN>`）
- Repository／Query の契約テストはローカルの PostgreSQL（`make db-up migrate-up`）に対して実行し、起動していなければ skip する。外部カタログのゲートウェイは偽の HTTP サーバに対してテストする
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## 環境変数（DB 以外）

| 環境変数 | 既定値（local） | 内容 |
|---|---|---|
| `ADMIN_TOKEN` | `local-admin-token` | 書き込み系 API のトークン（既定値は開発用） |
| `OPENBD_BASE_URL` | `https://api.openbd.jp` | openBD（登録・キー不要） |
| `RAKUTEN_APPLICATION_ID` / `RAKUTEN_ACCESS_KEY` | なし | 楽天ウェブサービスのアプリ ID とアクセスキー（任意。両方あるときだけ書影を楽天で補う） |
| `RAKUTEN_BASE_URL` | `https://openapi.rakuten.co.jp` | 楽天ウェブサービス |

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 書誌・書影 | openBD（登録不要）＋楽天ブックス書籍検索 API（任意） |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。
