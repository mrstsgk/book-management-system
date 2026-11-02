# Frontend

React + Vite + TypeScript（public・admin）。方針は [architecture.md](./architecture.md)。

## 前提

```bash
mise install   # Node 26.13.0 / pnpm 12.16.1
cd frontend
pnpm install
```

## コマンド（workspace ルート）

```bash
pnpm dev              # public :3000
pnpm dev:admin        # admin :3001
pnpm test            # 全アプリ
pnpm typecheck       # 全アプリ
pnpm lint            # ESLint（依存の向き・../ の相対 import・default export の禁止も検査。architecture.md §12）
pnpm format          # Prettier で整形（CI は pnpm format:check）
pnpm gen:api         # Orval で API クライアント・MSW ハンドラ・zod を全アプリで再生成（各アプリの orval.config.ts が ../openapi/dist/<channel>.yaml を読む。YAML を変えたら先に make -C ../backend openapi）
pnpm build
pnpm storybook        # :6006
pnpm build-storybook
```

## 管理画面のログイン

管理の API（`make run-admin`）を `ADMIN_ID` / `ADMIN_PASSWORD_HASH` を設定して起動し（手順は [`backend/README.md`](../backend/README.md)）、`http://localhost:3001/login` から ID とパスワードでログインする（脅威ごとの対策は [管理画面の概要設計](../docs/design/admin-screens.md)）。

Vite のプロキシは Host を書き換えないので、開発サーバーを普通に開けば、バックエンドの書き込み時の Origin と Host の一致確認を通る。

## E2E テスト（Playwright、ローカル実行のみ）

本のライフサイクル・分野タグのライフサイクルの重要導線だけを、実バックエンド・実 DB・実フロントエンドサーバーに対して確認する（CI に載せない理由は [`architecture.md`](./architecture.md) §15）。**CI では実行しない。** 事前に次を起動しておく。

```bash
# 1. バックエンド（別ターミナルで 2 つ）
cd backend
make db-up && make migrate-up
make run-public          # :8080
make run-admin           # :8081

# 2. ログイン情報（バックエンドの ADMIN_ID と、ADMIN_PASSWORD_HASH の元のパスワード）。
#    E2E の helpers.ts の login() が読む。ハッシュの作り方は backend/README.md
export E2E_ADMIN_ID=admin
export E2E_ADMIN_PASSWORD='...'

# 3. フロントエンド開発サーバー（別ターミナルで 2 つ）
cd frontend
pnpm dev                  # public :3000
pnpm dev:admin            # admin :3001

# 4. E2E を実行
cd frontend
pnpm test:e2e
```

初回だけ Chromium のダウンロードが要る（`npx playwright install chromium`）。

## 構成

- `public/` … 読んだ本の紹介（誰でも見られる。スマホ主）
- `admin/` … 自分用のログイン・本と分野タグの登録・編集（PC のみ。デジタル庁 DS）
- `packages/ui/` … デジタル庁 DS（Tailwind テーマ + 取り込み部品。現状 Button）。拠り所が DS の admin 用。public は Apple HIG に沿う（public は DS から移行中。`architecture.md` §4）
