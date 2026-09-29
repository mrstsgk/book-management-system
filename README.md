# Book Management System

自分が読んだ本を、一言まとめ・感想・評価・分野タグを添えて紹介するポートフォリオです。バックエンド API（Go / Echo）と Web 画面（React / Vite）を1つのリポジトリで作っています。

<!-- スクショ: 公開画面（一覧・詳細・分野別の集計）の実装後に、ここへ画像を差し込む -->
> 画面のスクショは、公開画面を実装したあとに追加します。

このリポジトリで伝えたいことは次の3つです。何を伝えるために作っているかは [`docs/requirements.md`](./docs/requirements.md) に書いています。

- 読んだ本と感想から、学び続けていること
- 設計・実装・テスト・CI から、設計と実装を丁寧に進めていること
- ADR・`CLAUDE.md`・hooks・PR の進め方から、AI とどう協業し、そのためのハーネスをどう作っているか

## 何ができるか

| 対象 | できること | 状況 |
|---|---|---|
| API | 本の登録・更新・削除（ISBN から書誌を openBD、書影を openBD／楽天ブックスで取得）。一覧のキーワード検索と分野タグでの絞り込み、分野タグごとの冊数、楽天由来の情報の期限管理、起動時の見本データ投入 | 実装済み（API の一覧は [`backend/architecture.md`](./backend/architecture.md#プロダクト-api-範囲)） |
| 公開画面 | 一覧（検索・絞り込み・ページング）、詳細、分野別の集計。誰でも見られる | 実装中 |
| 管理画面 | ログイン、本の登録・編集・削除、タグの管理。自分だけが使う | 未着手 |

満たすべき要件は [`docs/specifications.md`](./docs/specifications.md) にまとめています。

## 設計の方針

| 項目 | 方針 | 詳しくは |
|---|---|---|
| バックエンドの層 | オニオンアーキテクチャ（Presentation → UseCase → Domain ← Infrastructure）。Domain は Echo・GORM・SQL・外部 API に依存しない | [`backend/architecture.md` §2](./backend/architecture.md#2-層と依存) |
| 読み書きの分離 | CQRS。書き込みは Repository と Command DTO、読み取りは Read Model を返す Query に分ける（DB は1つ） | [`backend/architecture.md` §2](./backend/architecture.md#cqrsrepository-と-query) |
| API の契約 | Go の DTO と Handler を正とし、swag で OpenAPI を出力。フロントは Orval で TypeScript の型とクライアントを生成する。生成物は手で直さず、CI で再生成してずれがあれば落とす | [`docs/architecture.md` §4](./docs/architecture.md#4-openapi--型契約境界) |
| 外部カタログ | 書誌は openBD だけから取る。書影は openBD に無ければ楽天で補い、楽天の規約（保持は最長3か月）に合わせて期限を管理する | [ADR](./docs/adr/2026-09-28-rakuten-for-cover-only.md)・[設計](./docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md) |
| フロントエンド | React + Vite + TypeScript。UI はデジタル庁デザインシステム。サーバー状態は TanStack Query、表示とロジックは hooks とコンポーネントに分ける | [`frontend/architecture.md`](./frontend/architecture.md) |

## AI との協業とハーネス

開発は Claude Code と進めています。AI に任せる範囲を広げるほど、ルールを「読めば守れる」形にし、機械的に確かめられる部分は hook で強制するようにしています。

**指示の置き場所**

- [`CLAUDE.md`](./CLAUDE.md)（AI 向けの入口。英語）が、日本語で書いた方針とルール（`docs/`）を参照する。同じ内容を二重に持たない
- Cursor 向けに同じ入口を [`.cursor/rules/`](./.cursor/rules/) に置く
- 「ソースは How、テストは What、コミットは Why、コメントは Why not」のように、何をどこに書くかを決めている

**hooks による強制**（[`.claude/hooks/`](./.claude/hooks/)、[`.claude/settings.json`](./.claude/settings.json) で登録）

| hook | タイミング | 何を確かめるか |
|---|---|---|
| `require-tests-on-change.sh` | Stop（ブロック） | ソースを変えたのに、同じ場所に同名のテストの変更が無い |
| `require-adr-on-decision.sh` | Stop（ブロック） | 依存・インフラのファイル（`go.mod`・`package.json`・CI 設定など）を変えたのに ADR が無い |
| `require-db-docs-on-migration.sh` | Stop（ブロック） | マイグレーションを変えたのに DB の資料（AI 向け JSON・人向け ER 図）が更新されていない |
| `adr-reminder.sh` | デバッグ・レビュー系スキルの完了後（ブロックなし） | 根本原因の判断や見送ったレビュー指摘を ADR に残すよう促す |
| `precompact-save-evidence.sh` / `postcompact-load-evidence.sh` | 会話の圧縮の前後 | 圧縮前の作業ツリーの状態を保存し、圧縮後に読み戻す |

**進め方**

- 要求（[`requirements.md`](./docs/requirements.md)）→ 要件（[`specifications.md`](./docs/specifications.md)）→ 機能ごとの設計（[`docs/superpowers/specs/`](./docs/superpowers/specs/)）→ 実装計画（[`docs/superpowers/plans/`](./docs/superpowers/plans/)）→ TDD で実装、の順に進める
- 独立した機能は、AI のエージェントを別々の git worktree で並行に動かし、それぞれ PR にする
- PR は CodeRabbit（[`.coderabbit.yaml`](./.coderabbit.yaml)）がレビューし、指摘は1件ごとにコミットして、そのスレッドへコミット ID を添えて返信する（[PR のルール](./docs/rules/git-pull-request.md)）
- 自明でない判断は ADR に残す。採用しなかった案や、見送ったレビュー指摘も残す（[`docs/adr/`](./docs/adr/)。例: [結合テストを足す判断と、CI に DB を載せない判断](./docs/adr/2026-09-29-presentation-infra-integration-tests.md)）

## 品質

**テスト**（[テストのルール](./docs/rules/testing.md)。カバレッジの数値目標は置かない）

| 層 | 何で確かめるか |
|---|---|
| Domain | VO・Entity の単体テスト（不変条件・境界値） |
| UseCase | 手書きの Fake を注入したテスト（DB に繋がない） |
| Handler | `httptest` で実リクエストを送り、入力エラーとドメインのエラーが HTTP ステータスに変わることを確かめる |
| Repository／Query／外部ゲートウェイ | 実際の PostgreSQL や外部サービスの mock に対する契約テスト |
| Presentation → Infrastructure | Handler ごとに、返しうる HTTP ステータスごとの結合テスト（実 DB。ローカルで実行し、CI では DB が無いため skip） |
| フロントエンド | Vitest + React Testing Library。API の境界は MSW で差し替える |

**CI**（[`.github/workflows/`](./.github/workflows/)）

| | 実行するもの |
|---|---|
| Backend CI | gofmt、go vet、golangci-lint（循環的複雑度の上限10など）、govulncheck、テスト、ビルド、OpenAPI のずれの検査 |
| Frontend CI | 依存の脆弱性検査、OpenAPI から生成した型のずれの検査、型検査、ESLint、Prettier、テスト、ビルド、Storybook のビルド |

## セットアップ

前提: [mise](https://mise.jdx.dev/)・Docker（**Engine 28.0.0 以上**。それ未満では `127.0.0.1` に限定した公開ポートでも同じネットワークの他ホストから到達できる場合があるため）

```bash
mise install              # Go / Node / pnpm（版は .mise.toml）

# バックエンド
cd backend
make tools
make db-up                # PostgreSQL
make migrate-up
make run                  # http://localhost:8080/health
# 書き込み系 API は Authorization: Bearer local-admin-token（ローカルの既定値）。書影を楽天で補う場合は RAKUTEN_APPLICATION_ID / RAKUTEN_ACCESS_KEY を設定（backend/README.md）

# フロントエンド（別ターミナル）
cd frontend
pnpm install
pnpm dev                  # http://localhost:3000
```

詳細は [`backend/README.md`](./backend/README.md) / [`frontend/README.md`](./frontend/README.md)。

```
backend/    # Go / Echo API（PostgreSQL + GORM / golang-migrate / swag。書誌・書影は openBD / 楽天ブックス API）
frontend/   # React + Vite + TypeScript（pnpm workspace: web, packages/ui）
docs/       # 方針・ルール・ADR
```

## docs の案内

| 知りたいこと | 文書 |
|---|---|
| 誰に何を伝えるために作るか | [`docs/requirements.md`](./docs/requirements.md) |
| システムと開発の進め方が満たすこと | [`docs/specifications.md`](./docs/specifications.md) |
| リポジトリ全体の構成と契約の境界 | [`docs/architecture.md`](./docs/architecture.md) |
| バックエンド／フロントエンドの設計 | [`backend/architecture.md`](./backend/architecture.md) / [`frontend/architecture.md`](./frontend/architecture.md) |
| DB のテーブル構成 | [`docs/db/backend-schema.md`](./docs/db/backend-schema.md) |
| 判断の記録 | [`docs/adr/`](./docs/adr/) |
| 機能ごとの設計と実装計画 | [`docs/superpowers/specs/`](./docs/superpowers/specs/) / [`docs/superpowers/plans/`](./docs/superpowers/plans/) |
| 開発のルール | [ブランチ](./docs/rules/git-branch.md)・[コミット](./docs/rules/git-commit-message.md)・[PR](./docs/rules/git-pull-request.md)・[テスト](./docs/rules/testing.md)・[ADR](./docs/rules/adr.md)・[DB 資料](./docs/rules/db-documentation.md) |
