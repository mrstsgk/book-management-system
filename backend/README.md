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
make run            # 本が1冊も無ければ、起動時に見本データ（5冊）が入る
# http://localhost:8080/health
# http://localhost:8080/api/books
# 登録: curl -H "Authorization: Bearer local-admin-token" -H "Content-Type: application/json" \
#   -d '{"isbn":"9784873118703","summary":"要約","tagIds":[],"comment":"感想","rating":5}' http://localhost:8080/api/books

# OpenAPI 再排出（DTO/Handler 変更後）
make swagger
```

## 現状

- 読んだ本の API（一覧・詳細・登録・更新・削除）とカタログの確認 API、分野タグの API（一覧・追加・改名・削除）（範囲は `architecture.md` §4）
- 書誌は ISBN で openBD から、書影は Google Books → openBD の順で取得する（Google Books は `GOOGLE_BOOKS_API_KEY` を設定したときだけ）。起動時に1回、書影の無い本の書影を探す
- 書き込み系は管理者トークンが必要（`Authorization: Bearer <ADMIN_TOKEN>`）
- Repository／Query の契約テストはローカルの PostgreSQL（`make db-up migrate-up`）に対して実行し、起動していなければ skip する。外部カタログのゲートウェイは偽の HTTP サーバに対してテストする
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## 環境変数（DB 以外）

| 環境変数 | 既定値 | 内容 |
|---|---|---|
| `ADMIN_TOKEN` | `local-admin-token` | 書き込み系 API のトークン（既定値は開発用） |
| `OPENBD_BASE_URL` | `https://api.openbd.jp` | openBD（登録・キー不要） |
| `GOOGLE_BOOKS_API_KEY` | なし | Google Books API のキー（任意。Google Cloud で Books API を有効にして発行する）。設定したときだけ書影を Google Books から取る。キー無しの呼び出しは全員で共有する日次枠を使い切られているので使わない |
| `GOOGLE_BOOKS_BASE_URL` | `https://www.googleapis.com` | Google Books API（https のみ。キーをクエリで送るため） |

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 書誌・書影 | 書誌は openBD（登録不要）、書影は Google Books（キー任意）→ openBD |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。
