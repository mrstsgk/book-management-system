# Book Management System

書籍・著者を管理するシステム。バックエンド API（Go / Echo）と管理画面（React / Vite）のモノレポ。

方針の正は [`docs/architecture.md`](./docs/architecture.md)（バックエンド詳細は [`backend/architecture.md`](./backend/architecture.md)、フロントエンド詳細は [`frontend/architecture.md`](./frontend/architecture.md)）。

## 構成

```
backend/    # Go / Echo API（PostgreSQL + GORM / golang-migrate / swag）
frontend/   # React + Vite + TypeScript（pnpm workspace: web, packages/ui）
docs/       # 方針・ルール・ADR
```

## セットアップ

前提: [mise](https://mise.jdx.dev/)・Docker

```bash
mise install              # Go / Node / pnpm（版は .mise.toml）

# バックエンド
cd backend
make tools
make db-up
make migrate-up
make run                  # http://localhost:8080/health

# フロントエンド（別ターミナル）
cd frontend
pnpm install
pnpm dev                  # http://localhost:3000
```

詳細は [`backend/README.md`](./backend/README.md) / [`frontend/README.md`](./frontend/README.md)。

## 開発ルール

- ブランチ: [`docs/rules/git-branch.md`](./docs/rules/git-branch.md)
- コミットメッセージ: [`docs/rules/git-commit-message.md`](./docs/rules/git-commit-message.md)
- プルリクエスト: [`docs/rules/git-pull-request.md`](./docs/rules/git-pull-request.md)
- テスト: [`docs/rules/testing.md`](./docs/rules/testing.md)
- 判断記録（ADR）: [`docs/rules/adr.md`](./docs/rules/adr.md)
