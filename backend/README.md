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
make run            # 本が1冊も無ければ、起動時に見本データ（10冊）が入る
# http://localhost:8080/health
# http://localhost:8080/api/books
# ログイン: curl -c cookie.txt -H "Origin: http://localhost:8080" -H "Content-Type: application/json" \
#   -d '{"id":"admin","password":"..."}' http://localhost:8080/api/auth/login
# 登録: curl -b cookie.txt -H "Origin: http://localhost:8080" -H "Content-Type: application/json" \
#   -d '{"isbn":"9784873118703","summary":"要約","tagIds":[],"comment":"感想","rating":5}' http://localhost:8080/api/books

# OpenAPI 再排出（DTO/Handler 変更後）
make swagger
```

## 現状

- 読んだ本の API（一覧・詳細・登録・更新・削除）とカタログの確認 API、分野タグの API（一覧・追加・改名・削除）（範囲は `architecture.md` §4）
- 書誌と書影は ISBN で openBD から取得する
- 書き込み系はログイン済みのセッション Cookie が必要
- Repository／Query の契約テストはローカルの PostgreSQL（`make db-up migrate-up`）に対して実行し、起動していなければ skip する。外部カタログのゲートウェイは偽の HTTP サーバに対してテストする
- OpenAPI: `make swagger` → `backend/api/docs/`（手編集禁止。CI でドリフト検知）

## 環境変数（DB 以外）

| 環境変数 | 既定値 | 内容 |
|---|---|---|
| `ADMIN_ID` | なし | 管理画面のログイン ID。未設定ならログインできない |
| `ADMIN_PASSWORD_HASH` | なし | パスワードの bcrypt ハッシュ。`echo -n 'password' \| go run ./cmd/hashpw` で作る |
| `OPENBD_BASE_URL` | `https://api.openbd.jp` | openBD（登録・キー不要） |

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + validator + swag |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 書誌・書影 | openBD（登録不要） |
| FE 契約 | swag 排出 OpenAPI → TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。
