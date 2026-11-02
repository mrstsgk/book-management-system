# Book Management System

自分が読んだ本を、一言まとめ・感想・評価・分野タグを添えて紹介するポートフォリオです。バックエンド API（Go / Echo）と Web 画面（React / Vite）を1つのリポジトリで作っています。

このリポジトリで伝えたいことは次の3つです。何を伝えるために作っているかは [`docs/requirements.md`](./docs/requirements.md) に書いています。

- 読んだ本と感想から、学び続けていること
- 設計・実装・テスト・CI から、設計と実装を丁寧に進めていること
- ADR・`CLAUDE.md`・hooks・PR の進め方から、AI とどう協業し、そのためのハーネスをどう作っているか

本書は資料の**入口**で、読む順番と各資料の役割だけを書く。規則の中身は各資料が正で、ここには書き写さない（各資料の管轄は [`docs/architecture.md`「文書の言語」](./docs/architecture.md#文書の言語) の表）。

## 何ができるか

| 対象 | できること | 状況 |
|---|---|---|
| API | 本の登録・更新・削除（ISBN から書誌と書影を openBD で取得）。一覧のキーワード検索と分野タグでの絞り込み、分野タグごとの冊数、起動時の見本データ投入 | 実装済み（API の一覧は [`backend/architecture.md`](./backend/architecture.md#プロダクト-api-範囲)） |
| 公開画面 | 一覧（検索・絞り込み・ページング）、詳細、分野別の集計。誰でも見られる | 実装中 |
| 管理画面 | ログイン、本の登録・編集・削除、タグの管理。自分だけが使う | 実装済み |

## 1. 全員が最初に読む

| 順 | 資料 | 何が書いてあるか |
|---|---|---|
| 1 | [`docs/architecture.md`](./docs/architecture.md) | リポジトリ全体の方針、バックエンド／フロントエンドの境界、API 契約（swag → Orval）の流れ |
| 2 | [`docs/rules/development-flow.md`](./docs/rules/development-flow.md) | 要求・要件 → 設計 → 実装（TDD）→ PR → リリースの流れ |
| 3 | [`docs/rules/git-branch.md`](./docs/rules/git-branch.md)・[`git-commit-message.md`](./docs/rules/git-commit-message.md)・[`git-pull-request.md`](./docs/rules/git-pull-request.md) | ブランチ・コミット・PR の規則（レビュー分類を含む） |
| 4 | [`docs/rules/testing.md`](./docs/rules/testing.md) | テストの書き方・置き場所 |
| 5 | [`docs/rules/code-comments.md`](./docs/rules/code-comments.md) | コードのコメントの書き方 |

## 2. 担当に応じて読む

| 担当 | 資料 | 何が書いてあるか |
|---|---|---|
| 要求・要件 | [`docs/requirements.md`](./docs/requirements.md) | 誰に何を伝えるために作るか |
| | [`docs/specifications.md`](./docs/specifications.md) | システムと開発の進め方が満たすこと |
| 設計 | [`docs/superpowers/specs/`](./docs/superpowers/specs/) | 機能ごとの設計 |
| | [`docs/design/`](./docs/design/) | 機能ごとの概要設計 `docs/design/<機能>.md`（テンプレート `_template.md`）と、画面の HTML モック（公開画面・管理画面） |
| | [`docs/adr/`](./docs/adr/)・[`docs/rules/adr.md`](./docs/rules/adr.md) | 自明でない判断の記録と、その残し方 |
| バックエンド | [`backend/README.md`](./backend/README.md) | ローカルでの起動とコマンド |
| | [`backend/architecture.md`](./backend/architecture.md) | 層・命名・HTTP／OpenAPI（swag）・ディレクトリ |
| | [`docs/rules/db-documentation.md`](./docs/rules/db-documentation.md) | migration を変えたときの DB 資料の更新 |
| フロントエンド | [`frontend/README.md`](./frontend/README.md) | ローカルでの起動とコマンド・E2E の実行 |
| | [`frontend/architecture.md`](./frontend/architecture.md) | 設計・ディレクトリ・命名・デジタル庁 DS の使い方 |
| API 契約 | [`docs/architecture.md`](./docs/architecture.md) §4 | Go の DTO／Handler を正とし、swag で排出した OpenAPI から Orval で TypeScript を生成する流れ |

## 3. 必要になったら引く

| 資料 | 何が書いてあるか |
|---|---|
| [`docs/db/backend-schema.md`](./docs/db/backend-schema.md) | DB のテーブル構成（ER 図。AI 向けの JSON は同じディレクトリ） |
| [`docs/adr/`](./docs/adr/) | 判断の記録（採用しなかった案・見送ったレビュー指摘も残す） |
| [`docs/superpowers/plans/`](./docs/superpowers/plans/) | 機能ごとの実装計画 |

## 環境構築（ローカル）

前提: [mise](https://mise.jdx.dev/)（Go / Node / pnpm の版を `.mise.toml` で固定）・Docker（**Engine 28.0.0 以上**。それ未満では `127.0.0.1` に限定した公開ポートでも同じネットワークの他ホストから到達できる場合があるため）

```bash
mise install              # Go / Node / pnpm（版は .mise.toml）

# バックエンド
cd backend
make tools
make db-up                # PostgreSQL（使えるようになるまで待つ）
make migrate-up
make run                  # http://localhost:8080/health
# 書き込み系 API はログイン（POST /api/auth/login）で得た Cookie が要る

# フロントエンド（別ターミナル）
cd frontend
pnpm install
pnpm dev                  # http://localhost:3000
```

詳細は [`backend/README.md`](./backend/README.md) / [`frontend/README.md`](./frontend/README.md)。

## ディレクトリ構成

各ディレクトリの中の構成は、右に書いた資料が正。

```text
.
├── backend/               # Go / Echo の API（backend/architecture.md）
│   ├── cmd/               #   実行ポイント（api）と手書き DI、見本データ
│   ├── internal/          #   presentation / usecase / domain / infrastructure（各層の下にドメイン別 + common）
│   ├── api/docs/          #   swag が排出した OpenAPI（手編集しない）
│   ├── migrations/        #   DB マイグレーション（SQL）
│   └── config/            #   設定の読み込み
├── frontend/              # React + Vite の Web（frontend/architecture.md）
│   ├── web/               #   読んだ本の紹介画面と、自分用の登録・編集画面
│   └── packages/          #   アプリ横断の共有（ui/ = デジタル庁 DS）
├── docs/                  # 人が読む資料（本書の 1〜3 を参照）
│   ├── rules/             #   開発の規則
│   ├── adr/               #   判断の記録
│   ├── db/                #   DB スキーマ資料
│   ├── design/            #   画面の HTML モック
│   └── superpowers/       #   機能ごとの設計（specs/）と実装計画（plans/）
├── .github/               # CI（workflows/）・CI 用スクリプト（scripts/）・PR テンプレート
├── .claude/ .cursor/      # AI エージェントの設定・hook・ルール
├── CLAUDE.md AGENTS.md    # AI エージェントの入口
├── .coderabbit.config.ts  # CodeRabbit の設定（レビュー分類ラベルで profile を切り替える）
└── .mise.toml             # Go・Node・pnpm の版の固定
```

## AI との協業とハーネス

開発は Claude Code と進めています。AI に任せる範囲を広げるほど、ルールを「読めば守れる」形にし、機械的に確かめられる部分は hook と CI で強制するようにしています。

**指示の置き場所**

- [`CLAUDE.md`](./CLAUDE.md)（Claude Code 向けの入口。英語）が、日本語で書いた方針とルール（`docs/`）を参照する。同じ内容を二重に持たない
- その他のエージェント向けの共通ルールは [`AGENTS.md`](./AGENTS.md)、Cursor 向けの入口は [`.cursor/rules/`](./.cursor/rules/)
- 「ソースは How、テストは What、コミットは Why、コメントは Why not」のように、何をどこに書くかを決めている（[`docs/rules/code-comments.md`](./docs/rules/code-comments.md)）

**hooks による強制**（[`.claude/hooks/`](./.claude/hooks/)、[`.claude/settings.json`](./.claude/settings.json) で登録）

| hook | タイミング | 何を確かめるか |
|---|---|---|
| `require-tests-on-change.sh` | Stop（ブロック） | ソースを変えたのに、同じ場所に同名のテストの変更が無い（AI の作業時の先行チェック。正は CI の `Test Presence`） |
| `require-adr-on-decision.sh` | Stop（ブロック） | 依存・インフラのファイル（`go.mod`・`package.json`・CI 設定など）を変えたのに ADR が無い |
| `require-db-docs-on-migration.sh` | Stop（ブロック） | マイグレーションを変えたのに DB の資料（AI 向け JSON・人向け ER 図）が更新されていない |
| `adr-reminder.sh` | Stop（ブロックなし） | 直前にデバッグ・レビュー系のスキルを使っていたら、根本原因の判断や見送ったレビュー指摘を ADR に残すよう促す |
| `precompact-save-evidence.sh` / `postcompact-load-evidence.sh` | 会話の圧縮の前後 | 圧縮前の作業ツリーの状態を保存し、圧縮後に読み戻す |

テストの書き忘れの正は CI の `Test Presence`（人が書いた PR も検査するため）。hook は AI の作業時に早く気づくための先行チェック。

**進め方**

- 要求（[`requirements.md`](./docs/requirements.md)）→ 要件（[`specifications.md`](./docs/specifications.md)）→ 機能ごとの設計（[`docs/superpowers/specs/`](./docs/superpowers/specs/)）→ 実装計画（[`docs/superpowers/plans/`](./docs/superpowers/plans/)）→ TDD で実装、の順に進める（[開発の流れ](./docs/rules/development-flow.md)）
- 独立した機能は、AI のエージェントを別々の git worktree で並行に動かし、それぞれ PR にする
- PR を作る前に AI がレビュー分類（設計・仕様 / 通常の実装 / 重要な実装 / レビュー不要）を提案し、人が選ぶ。選んだ分類のラベルで CodeRabbit（[`.coderabbit.config.ts`](./.coderabbit.config.ts)）の profile と観点が切り替わる。指摘は吟味してから1件ごとにコミットし、そのスレッドへコミット ID を添えて返信する（[PR のルール](./docs/rules/git-pull-request.md)）
- 自明でない判断は ADR に残す。採用しなかった案や、見送ったレビュー指摘も残す（[`docs/adr/`](./docs/adr/)）

## 品質

**テスト**（[テストのルール](./docs/rules/testing.md)。カバレッジの数値目標は置かない）

| 層 | 何で確かめるか |
|---|---|
| Domain | VO・Entity の単体テスト（不変条件・境界値） |
| UseCase | 手書きの Fake を注入したテスト（DB に繋がない） |
| Handler | `httptest` で実リクエストを送り、入力エラーとドメインのエラーが HTTP ステータスに変わることを確かめる |
| Repository／Query／外部ゲートウェイ | 実際の PostgreSQL や外部サービスの mock に対する契約テスト（Repository／Query は SQL が複雑なものを優先） |
| Presentation → Infrastructure | Handler ごとに、返しうる HTTP ステータスごとの結合テスト（実 DB。ローカルでも CI でも実行し、DB に繋がらなければ失敗にする） |
| フロントエンド | Vitest + React Testing Library。API の境界は MSW（Orval が生成したハンドラ）で差し替える |
| E2E | Playwright。本の登録・編集・削除、タグの追加・本への付与・分野別集計への反映・改名・削除の重要導線だけを実バックエンド・実 DB に対して確認（[ADR](./docs/adr/2026-09-30-introduce-playwright-e2e.md)）。**ローカル実行のみ、CI には含めない**。手順は [`frontend/README.md`](./frontend/README.md#e2e-テストplaywrightローカル実行のみ) |

**CI**（[`.github/workflows/`](./.github/workflows/)）

| | 実行するもの |
|---|---|
| Backend CI | gofmt、go vet、golangci-lint（gosec・modernize・循環的複雑度の上限10など）、govulncheck、PostgreSQL を起動して migration を当ててからテスト（DB を使うテストも実行）、ビルド、OpenAPI のずれの検査 |
| Frontend CI | 依存の脆弱性検査、OpenAPI から生成した型のずれの検査、型検査、ESLint（依存の向き・default export の禁止を含む）、Prettier、テスト、ビルド、Storybook のビルド |
| Test Presence | API の operation ごとに、それを実装・利用するコード（Handler・画面）のテストがあるか（無ければ失敗）。変更したソースと同じディレクトリのテストが変わっていなければ警告 |
| Review Policy | レビュー分類のラベルが1つだけ付いているか。CodeRabbit の設定（`.coderabbit.config.ts`）の型検査とテスト |

## AI エージェント

Claude Code は [`CLAUDE.md`](./CLAUDE.md)、その他のエージェントは [`AGENTS.md`](./AGENTS.md) を入口にする（AI 向けの入口で、本書とは別に保つ）。
