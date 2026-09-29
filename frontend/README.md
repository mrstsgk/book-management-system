# Frontend

React + Vite + TypeScript。方針は [architecture.md](./architecture.md)。

## 前提

```bash
mise install   # Node 24.11.0 / pnpm 10.11.0
cd frontend
pnpm install
```

## コマンド（workspace ルート）

```bash
pnpm dev              # web :3000
pnpm test
pnpm typecheck
pnpm lint            # ESLint
pnpm format          # Prettier で整形（CI は pnpm format:check）
pnpm gen:api         # Orval で API クライアント・MSW ハンドラを再生成（backend/api/docs/swagger.yaml が入力）
pnpm build
pnpm storybook        # :6006
pnpm build-storybook
```

## 管理画面の管理者トークン

管理画面（`/admin`）から書き込むには、バックエンドの `ADMIN_TOKEN` と同じ値を `web/.env.development.local`（git 管理外）に書いてから `pnpm dev` を起動する。

```bash
echo 'VITE_ADMIN_TOKEN=local-admin-token' > web/.env.development.local   # バックエンドの既定値
```

値は画面のコードに埋め込まれるので、この設定でビルドしたものを公開しない（[ADR](../docs/adr/2026-09-30-admin-token-from-env.md)）。

## 構成

- `web/` … 読んだ本の紹介画面と、自分用の登録・編集画面
- `packages/ui/` … デジタル庁 DS（Tailwind テーマ + 取り込み部品。現状 Button）
