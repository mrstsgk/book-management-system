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
- Repository／Query の契約テストはローカルの PostgreSQL（`make db-up migrate-up`）に対して実行し、起動していなければ skip する。外部カタログ（openBD・楽天）のゲートウェイは偽の HTTP サーバに対してテストする
- 書籍は ISBN を持てる。書籍の作成・更新時に ISBN で外部カタログを引き、書影の URL と提供元を保存する（画像そのものは保存しない）。Amazon のリンクは ISBN-10 から導出する
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 書誌・書影 | openBD（登録不要）。書影が無い本は楽天ブックス書籍検索 API で補う（任意） |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。

## 外部カタログ（書誌・書影）

| 環境変数 | 既定値 | 内容 |
|---|---|---|
| `OPENBD_BASE_URL` | `https://api.openbd.jp` | openBD（登録・キー不要） |
| `RAKUTEN_APPLICATION_ID` | なし | 楽天ウェブサービスのアプリ ID（任意） |
| `RAKUTEN_ACCESS_KEY` | なし | 楽天ウェブサービスのアクセスキー（任意） |
| `RAKUTEN_BASE_URL` | `https://openapi.rakuten.co.jp` | 楽天ウェブサービス |

楽天のアプリ ID とアクセスキーの両方を設定したときだけ、openBD に書影が無い本の書影を楽天ブックスで補う。未設定なら openBD だけを使う（書影が無い本は書影なし）。

- 書影は提供元の URL をそのまま表示する（保存・加工しない）。openBD・楽天とも利用は本の紹介目的に限られる
- 楽天の書影を表示する画面には楽天ウェブサービスのクレジット表示が必要（`coverSource` が `rakuten` のとき）

