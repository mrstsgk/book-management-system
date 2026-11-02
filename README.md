# book-management-system

自分が読んだ本を、一言まとめ・感想・評価・分野タグを添えて紹介するポートフォリオのモノレポ。誰でも見られる公開画面（public）と、自分だけが使う管理画面（admin）を、Go / Echo の API と React / Vite の Web で作っている。

開発に参加したら、まず本書を読む。本書は資料の**入口**で、読む順番と各資料の役割だけを書く。規則の中身は各資料が正で、ここには書き写さない（各資料の管轄は [`docs/architecture.md`「文書の言語」](docs/architecture.md#文書の言語) の表）。

## 1. 全員が最初に読む

| 順 | 資料 | 何が書いてあるか |
|---|---|---|
| 1 | [`docs/architecture.md`](docs/architecture.md) | リポジトリ全体の方針、チャネル（public / admin）の境界、API 契約（OpenAPI）の置き場と規則 |
| 2 | [`docs/rules/development-flow.md`](docs/rules/development-flow.md) | 要件定義 → 概要設計 → 詳細設計（OpenAPI）→ 実装（TDD）→ PR → リリースの流れ |
| 3 | [`docs/rules/git-branch.md`](docs/rules/git-branch.md)・[`git-commit-message.md`](docs/rules/git-commit-message.md)・[`git-pull-request.md`](docs/rules/git-pull-request.md) | ブランチ・コミット・PR の規則 |
| 4 | [`docs/rules/testing.md`](docs/rules/testing.md) | テストの書き方・置き場所 |
| 5 | [`docs/rules/code-comments.md`](docs/rules/code-comments.md) | コードのコメントの書き方 |

## 2. 担当に応じて読む

| 担当 | 資料 | 何が書いてあるか |
|---|---|---|
| 要件定義・モック | [`docs/requirements/PROCESS.md`](docs/requirements/PROCESS.md) | 要件定義の進め方。正本は同じディレクトリの `status-map.md` と GitHub Issue（ラベル `論点`） |
| | [`AGENTS.md`](AGENTS.md) の Mocks | `mock/` の HTML モックの作り方 |
| 概要設計 | [`docs/design/_template.md`](docs/design/_template.md) | 概要設計の書き方（1 機能 1 ファイル `docs/design/<機能>.md`） |
| バックエンド | [`backend/README.md`](backend/README.md) | ローカルでの起動とコマンド |
| | [`backend/architecture.md`](backend/architecture.md) | 層・命名・HTTP（OpenAPI からの生成）・ディレクトリ |
| | [`docs/rules/db-documentation.md`](docs/rules/db-documentation.md) | migration を変えたときの DB 資料の更新 |
| フロントエンド | [`frontend/README.md`](frontend/README.md) | ローカルでの起動とコマンド |
| | [`frontend/architecture.md`](frontend/architecture.md) | 設計・ディレクトリ・チャネル別のデザインの拠り所 |
| | [`docs/rules/ui-design.md`](docs/rules/ui-design.md)・[`input-characters.md`](docs/rules/input-characters.md) | 管理画面のデザインルール・入力文字の共通ルール |
| API 契約 | [`docs/architecture.md`](docs/architecture.md) §4 | `openapi/` の YAML の置き場・規則（コマンドは `make -C openapi help`） |

## 3. 必要になったら引く

| 資料 | 何が書いてあるか |
|---|---|
| [`docs/research/`](docs/research/README.md) | 論点の調査メモ（決定ではない） |
| [`docs/db/backend-schema.md`](docs/db/backend-schema.md) | DB のテーブル構成（ER 図） |

## 環境構築（ローカル）

前提: [mise](https://mise.jdx.dev/)（Go・Node・pnpm の版を `.mise.toml` で固定）と Docker（PostgreSQL 用。Engine 28.0.0 以上。それ未満では `127.0.0.1` に限定した公開ポートでも同じネットワークの他ホストから到達できる場合があるため）。

| 順 | やること | 手順の正 |
|---|---|---|
| 1 | リポジトリ直下で `mise install` | 本書 |
| 2 | DB を起動して migration を当て、API を起動する（public: `http://localhost:8080`、admin: `http://localhost:8081`） | [`backend/README.md`](backend/README.md)「ローカル開発」 |
| 3 | 依存を入れて Web を起動する（public: `http://localhost:3000`、admin: `http://localhost:3001`） | [`frontend/README.md`](frontend/README.md)「前提」「コマンド」 |

ローカルでは `.env` は不要（API は docker compose の DB を既定の接続先にし、Vite は `/api` を API へ中継する）。管理画面のログインに使う ID とパスワードのハッシュだけは環境変数で渡す（[`backend/README.md`](backend/README.md)）。

## ディレクトリ構成

各ディレクトリの中の構成は、右に書いた資料が正。

```text
.
├── backend/               # Go / Echo の API（backend/architecture.md）
│   ├── cmd/               #   実行ポイント（public-api・admin-api）と手書き DI
│   ├── internal/          #   業務領域別のコード（book/・tag/・auth/）と技術基盤（platform/）
│   ├── api/gen/           #   OpenAPI からの生成物（public/・admin/）
│   ├── migrations/        #   DB マイグレーション（SQL）
│   ├── config/            #   設定の読み込み
│   └── environments/      #   環境変数の例（.env.example）
├── frontend/              # React + Vite の Web（frontend/architecture.md）
│   ├── public/            #   公開画面（読んだ本の一覧・詳細・分野別の集計）
│   ├── admin/             #   管理画面（ログイン・本とタグの登録・編集・削除）
│   ├── packages/          #   アプリ横断の共有（ui/ = デジタル庁 DS）
│   └── environments/      #   環境変数の例（.env.example）
├── openapi/               # API 契約（docs/architecture.md §4）
│   ├── <area>/<channel>/  #   手書き YAML（業務領域 × チャネル）
│   ├── common/            #   チャネル共通のスキーマ
│   └── dist/              #   バンドル結果（public.yaml・admin.yaml）
├── docs/                  # 人が読む資料（本書の 1〜3 を参照）
│   ├── rules/             #   開発の規則
│   ├── requirements/      #   要件（機能群・論点）
│   ├── design/            #   概要設計（1 機能 1 ファイル）
│   ├── db/                #   DB スキーマ資料
│   └── research/          #   調査メモ
├── mock/                  # 要件確認用の HTML モック
├── .github/               # CI（workflows/）・CI 用スクリプト・PR テンプレート
├── .claude/               # AI エージェントの設定・hook
├── CLAUDE.md AGENTS.md    # AI エージェントの入口
└── .mise.toml             # Go・Node・pnpm の版の固定
```

デプロイはしない（要件）。本番イメージの形だけ `backend/Dockerfile.api` に置く。

## AI エージェント

Claude Code は [`CLAUDE.md`](CLAUDE.md)、その他のエージェントは [`AGENTS.md`](AGENTS.md) を入口にする（AI 向けの入口で、本書とは別に保つ）。
