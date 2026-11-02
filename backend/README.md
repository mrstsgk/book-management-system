# Backend

Go / Echo モノリス。方針の正は [`architecture.md`](./architecture.md)。

## ローカル開発

| 対象 | 方法 |
|---|---|
| Go | **ホスト**（版はリポジトリ直下 `.mise.toml` で固定） |
| DB | **Docker Compose**（PostgreSQL） |
| API コンテナ | 日常開発では使わない（本番イメージの形だけ。`Dockerfile.api`） |
| Dev Container | 使わない |

```bash
# リポジトリルート
mise install

cd backend
make db-up   # docker compose up -d --wait（DB が使えるようになるまで待つ）
# golang-migrate が必要: make tools（版は Makefile で固定。CI と同じ）
make migrate-up
make run-public   # 閲覧の API（:8080）。本が1冊も無ければ、起動時に見本データ（10冊）が入る
make run-admin    # 管理の API（:8081。別のターミナルで）
# http://localhost:8080/health
# http://localhost:8080/api/books

# API 契約（../openapi/ の YAML）を変更したら: lint・バンドル・BE 再生成
make openapi

# テスト（DB を使うテスト（契約・*_integration_test.go）は DB が無ければ失敗する）
make db-up && make migrate-up
make test
```

## 現状

- 構成: `internal/` の第一階層は業務領域別（`book`・`tag`・`auth`）。領域の中に Presentation → UseCase → Domain ← Infrastructure を置く（Repository／Query 分離）。技術基盤は `internal/platform/`、共通エラーは `internal/domainerr/`
- 閲覧（`cmd/public-api`）: 読んだ本の一覧・詳細、分野タグの一覧・冊数
- 管理（`cmd/admin-api`）: ログイン・ログアウト、本の登録・更新・削除とカタログの確認、分野タグの追加・改名・削除（範囲は `architecture.md` §4）
- 書誌と書影は ISBN で openBD から取得する
- OpenAPI: 正本は `../openapi/`（Design First）。`make openapi` で `api/gen/{public,admin}/` を生成（手編集禁止。CI でドリフト検知）

## 環境変数（DB 以外）

| 環境変数 | 既定値 | 内容 |
|---|---|---|
| `ADMIN_ID` | なし | 管理画面のログイン ID。未設定ならログインできない |
| `ADMIN_PASSWORD_HASH` | なし | パスワードの bcrypt ハッシュ。`read -rs PW; printf %s "$PW" \| go run ./cmd/hashpw; unset PW` で作る（`read -s` は入力を画面に出さず、shell の履歴にも残さない） |
| `OPENBD_BASE_URL` | `https://api.openbd.jp` | openBD（登録・キー不要） |

## スタック（要約）

| 項目 | 内容 |
|---|---|
| HTTP | Echo + oapi-codegen（Server IF・型）+ kin-openapi（入力検証） |
| 設計 | オニオン + DDD + CQRS（単一 DB） |
| DB | PostgreSQL + GORM / golang-migrate |
| 書誌・書影 | openBD（登録不要） |
| FE 契約 | 同じ `openapi/dist/` から Orval で TypeScript 生成（必須） |

詳細は [`architecture.md`](./architecture.md)。
