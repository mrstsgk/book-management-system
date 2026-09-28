# Book Management System

書籍・著者を管理するシステム。バックエンド API（Go / Echo）と管理画面（React / Vite）のモノレポ。

題材は小さな CRUD だが、このリポジトリの本当の目的は **「普段考えているソフトウェアアーキテクチャ」と「AI（Claude Code / Cursor）を開発に組み込むためのハーネス」を、動くコードと設定ファイルごと共有すること** にある。README はその全体像の地図で、詳細な方針は各文書（下記リンク）を正とする。

---

## 目次

1. [ソフトウェアアーキテクチャの考え方](#1-ソフトウェアアーキテクチャの考え方)
2. [AI 活用のハーネス](#2-ai-活用のハーネス)
3. [構成](#3-構成)
4. [セットアップ](#4-セットアップ)
5. [開発ルール](#5-開発ルール)

---

## 1. ソフトウェアアーキテクチャの考え方

### 1.1 根っこにある価値観: 変更容易性

設計判断はすべて **「変更の影響範囲を局所化し、関係ない機能を壊さない」** ために行う。そのための原則は次の 3 つに集約される。

| 原則 | 意味 |
|---|---|
| **目的駆動** | 名前・クラスは「存在」ではなく「目的」で切る。1 クラス 1 目的（例: 書籍の新規登録と更新は `CreateCommand` / `UpdateCommand` と UseCase を分ける） |
| **変更理由の分離 > 共通化** | 似ているだけの処理はまとめない。「同じ理由で変更されるか？」で判断し、`Common` / `Util` / `Manager` のような曖昧なクラスを作らない |
| **不正な状態を作れない設計** | 業務ルールは VO・Entity のコンストラクタ 1 か所で保証する。呼び出し側に同じ範囲チェックを散らさない |

### 1.2 バックエンド: オニオン + DDD 戦術 + CQRS

```text
presentation/http/{common,<domain>}      HTTP 変換のみ（業務ロジック・認可を持たない）
  → usecase/<domain>/{command,query}     シナリオ実行・認可
    → domain/{common,<domain>}           業務ルールの中核。他層に依存しない
       ↑
infrastructure/postgres/<domain>         Repository / Query 実装
infrastructure/gateway/<domain>          ExternalGateway 実装（openBD・楽天ブックスなど）
```

- **Domain は外部技術を知らない**: Echo・GORM・SQL・外部 SDK を Domain から参照しない。外との通信は Domain に置いた IF（Repository / Query / ExternalGateway）経由。IF の置き場を種類で変えず、常に Domain に置いて一貫させる
- **モデルは目的ごとに 3 つに分ける**: 同じ「書籍」でも Domain Model（`book.Book` + `Price` VO）・Read Model（`book.BookSummary`）・Persistence Model（`gorm` タグ付き `model`）は別物。DB テーブル構造を Domain Model にしない
- **CQRS は単一 DB で軽量に**: 書き込みは `Repository` + Command DTO、読み取りは `Query` + Read Model。DB 分離・Event Sourcing はしない
- **やらないことを明記する**: マイクロサービス・モジュラモノリス・全体 ES・Dev Container は採らない。境界が必要になったら後から育てる

正: [`backend/architecture.md`](./backend/architecture.md)

### 1.3 フロントエンド: bulletproof-react 寄せ + 責務での分離

- **Container/Presenter をフォルダ名で表現しない**。ロジックは `hooks/`、表示コンポーネントは props で描画に徹する、つなぎは `app/`。smart / dumb の分離は責務で行う
- **状態の置き場を種類で決める**: サーバ状態は TanStack Query、UI 状態は Zustand、横断注入は Context。サーバ由来データを二重保管しない
- **feature 間 import 禁止**・一方向依存（shared ← features ← app）
- **UI はデジタル庁デザインシステム**を `packages/ui` に取り込んで使う
- **Storybook は Presentational の状態カタログ**（ページ丸ごとは載せない、視覚回帰はしない）

正: [`frontend/architecture.md`](./frontend/architecture.md)

### 1.4 契約: Go 先行 → OpenAPI 排出 → TS 生成

```text
Go DTO + validate + 薄い Handler + swag
  → swag init（OpenAPI 排出）
  → Orval で TanStack Query フック・型・MSW ハンドラを生成
  → CI: 再生成して git diff --exit-code（ドリフトがあれば fail）
```

手書き OpenAPI を正にする方式は採らない。生成物は手編集しない。FE で API 型を手書きしない。

正: [`docs/architecture.md`](./docs/architecture.md) §4

### 1.5 「何をどこに書くか」

コード・テスト・コミット・コメントそれぞれに書く内容を分けている。情報が重複せず、どこを見れば何が分かるかが明確になる。

| 場所 | 書くこと |
|---|---|
| ソースコード | **How**（仕組み・ロジック） |
| テストコード | **What**（期待する振る舞い・仕様） |
| コミットログ | **Why**（変更の動機） |
| コードコメント | **Why not**（却下した代替案・自明でない制約） |
| ADR（`docs/adr/`） | コードにもコミットにも収まらない **判断の経緯**（採用も不採用も） |

例外として、Repository と複雑なメソッドには日本語で 1 行の What コメントを付ける（詳細は [`CLAUDE.md`](./CLAUDE.md)）。

---

## 2. AI 活用のハーネス

### 2.1 考え方

AI に「ルールを読んで守ってね」と頼むだけでは、セッションが変わる・コンテキストが圧縮されるたびに守られなくなる。そこで次の 3 段構えにしている。

| 段 | 手段 | 役割 |
|---|---|---|
| **① 伝える** | `CLAUDE.md` / `.cursor/rules/*.mdc` | AI の入口。方針の本文は持たず、日本語の正を `@import` / 参照するだけ |
| **② 強制する** | Claude Code hooks（Stop でブロック） | 機械的に判定できるルールは、守られるまで作業を終わらせない |
| **③ 検証する** | CI・レビュー（Copilot / CodeRabbit / サブエージェント） | 人と別の AI の目で最終確認する |

設計上のこだわり:

- **文書は MECE**: リポジトリ境界は `docs/architecture.md`、BE は `backend/architecture.md`、FE は `frontend/architecture.md`。同じことを 2 か所に書かない（片方だけ更新されて矛盾するため）
- **言語を使い分ける**: 人が読む方針は日本語、AI の入口は英語。長い方針を英語側に複製しない
- **Claude Code と Cursor で同じルールを共有**: `CLAUDE.md` と `.cursor/rules/*.mdc`（`alwaysApply: true`）が同じ日本語文書を指す
- **hook は「構造」だけを見る**: 中身の妥当性は grep では判定できないので、そこは人 / AI レビューに任せると割り切る。誤検知時のバイパス環境変数を必ず用意する

### 2.2 hooks（`.claude/hooks/`）

設定は [`.claude/settings.json`](./.claude/settings.json)。

| hook | イベント | 内容 | バイパス |
|---|---|---|---|
| `require-tests-on-change.sh` | Stop（ブロック） | ソースを変更したのに同じディレクトリ・同名のテスト（`*_test.go` / `*.test.ts(x)`）が変わっていなければ終了させない | `SKIP_TEST_CHECK=1` |
| `require-adr-on-decision.sh` | Stop（ブロック） | 依存・インフラ（`go.mod` / `package.json` / workflow など）を変えたのに `docs/adr/` が無い、または ADR の必須節が空なら終了させない | `SKIP_ADR_CHECK=1` |
| `require-db-docs-on-migration.sh` | Stop（ブロック） | `backend/migrations/*.sql` を変えたのに `docs/db/backend-schema.{json,md}` の両方が更新されていなければ終了させない | `SKIP_DB_DOCS_CHECK=1` |
| `adr-reminder.sh` | Stop（非ブロック） | デバッグ・コードレビュー系スキルを使ったターンの終わりに、根本原因の判断や指摘の見送りを ADR に残すよう促す（diff では判断の有無を検知できないため、ブロックはしない） | — |
| `precompact-save-evidence.sh` | PreCompact | `/compact` 直前の `git status` / diff を `.claude/_evidence/` に退避 | — |
| `postcompact-load-evidence.sh` | SessionStart（compact） | 退避した差分を圧縮後のコンテキストへ再注入し、「さっき何を変えていたか」を失わない | — |

### 2.3 その他の仕掛け

| 仕掛け | 場所 | 内容 |
|---|---|---|
| サブエージェント | [`.claude/agents/spec-drift-reviewer.md`](./.claude/agents/spec-drift-reviewer.md) | 実装が設計ドキュメントからずれていないかだけを見るレビュアー。コード品質レビュー（`/code-review`）とは役割を分ける |
| 出力スタイル | [`.claude/output-styles/Concise.md`](./.claude/output-styles/Concise.md) | 短く答える。**事実と仮説**、**提案と決定事項**を分けて書かせる |
| 権限 | `.claude/settings.json` の `permissions` | `.env`・秘密鍵の読み書きと `rm -rf` / `sudo` を deny、`git reset` / force push などは ask |
| ステータスライン | [`.claude/statusline-command.sh`](./.claude/statusline-command.sh) | モデル・effort・コンテキスト使用率を常時表示 |
| AI レビュー | [`.coderabbit.yaml`](./.coderabbit.yaml)・GitHub Copilot | 日本語でレビュー。ブランチ・コミット規約もレビュー観点に含める。指摘対応は 1 指摘 1 コミット + スレッドにコミット ID を返信 |
| ADR | [`docs/adr/`](./docs/adr/) | ライブラリ選定・見送ったレビュー指摘・根本原因の判断など、「なぜ A でなく B か」を残す（A = Architecture ではなく **Any**） |
| DB スキーマ資料 | [`docs/db/`](./docs/db/) | AI 向け JSON（型・nullable・意味）と人向け Mermaid ER 図の 2 本立て。AI がマイグレーションを合成し直さずに済む |

### 2.4 CI（`.github/workflows/`）

hook はローカルの AI 作業を縛るもの、CI はそれをすり抜けた変更を止める最後の砦。

- **Backend**: `gofmt` / `go vet` / `golangci-lint` / `govulncheck` / test / build / **OpenAPI ドリフト検知**
- **Frontend**: `pnpm audit` / **TS 生成ドリフト検知** / typecheck / lint / format / test / build / build-storybook

### 2.5 自分のリポジトリに持ち込むなら

1. `CLAUDE.md` に「What goes where」と文書の在りか（`@import`）だけを書く
2. 機械的に判定できるルール（テスト同梱・ADR・スキーマ資料）を Stop hook にし、バイパス変数を付ける
3. 判定できないルール（判断の記録など）は非ブロックのリマインダーに留める
4. 契約（OpenAPI など）は生成 + CI の diff チェックで AI の手書きを防ぐ

---

## 3. 構成

```
backend/    # Go / Echo API（PostgreSQL + GORM / golang-migrate / swag。書誌・書影は openBD / 楽天ブックス API）
frontend/   # React + Vite + TypeScript（pnpm workspace: web, packages/ui）
docs/       # 方針・ルール・ADR・DB スキーマ資料
.claude/    # Claude Code のハーネス（hooks / agents / output-styles / settings）
.cursor/    # Cursor 用ルール（CLAUDE.md と同じ文書を参照）
.github/    # CI・PR テンプレート
```

## 4. セットアップ

前提: [mise](https://mise.jdx.dev/)・Docker・`jq`（hooks が使用）

```bash
mise install              # Go / Node / pnpm（版は .mise.toml）

# バックエンド
cd backend
make tools
make db-up                # PostgreSQL
make migrate-up
make run                  # http://localhost:8080/health
# 書影を楽天ブックスで補う場合は RAKUTEN_APPLICATION_ID / RAKUTEN_ACCESS_KEY を設定する（任意。backend/README.md 参照）

# フロントエンド（別ターミナル）
cd frontend
pnpm install
pnpm dev                  # http://localhost:3000
```

詳細は [`backend/README.md`](./backend/README.md) / [`frontend/README.md`](./frontend/README.md)。

## 5. 開発ルール

- ブランチ: [`docs/rules/git-branch.md`](./docs/rules/git-branch.md)
- コミットメッセージ: [`docs/rules/git-commit-message.md`](./docs/rules/git-commit-message.md)
- プルリクエスト: [`docs/rules/git-pull-request.md`](./docs/rules/git-pull-request.md)
- テスト: [`docs/rules/testing.md`](./docs/rules/testing.md)
- 判断記録（ADR）: [`docs/rules/adr.md`](./docs/rules/adr.md)
- DB スキーマ資料: [`docs/rules/db-documentation.md`](./docs/rules/db-documentation.md)
