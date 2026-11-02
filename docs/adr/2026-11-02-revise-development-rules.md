# 開発規則・レビュー・CI の見直し

**日付:** 2026-11-02
**状態:** 採用
**カテゴリ:** review
**参照:** [PR #113](https://github.com/mrstsgk/book-management-system/pull/113)

## 背景・経緯

これまでの規則は AI の作業時に動く Stop hook（テストの有無・ADR・DB 資料）を中心に強制してきたが、hook は人が手で書いた PR を検査できない。また次の点が目立ってきた。

- CodeRabbit のレビューの強度（profile）が PR の性質（設計・通常の実装・重要な実装）によらず一定で、設計だけの Draft でも未実装を指摘される
- テスト規約が正常系の書き方に偏り、異常系（不正な入力・競合・外部障害）の観点を洗い出す手順が無かった
- DB を使うテスト（契約・結合・`cmd` の配線）が CI で skip され、配線ミスがマージ後まで見つからない
- 命名・コメント・依存の向きの規則が各文書に散らばり、機械的に検査されていない

## 決定

次を採用する。

| 項目 | 内容 |
|---|---|
| レビュー分類 | PR ごとに人が `review:design` / `review:standard` / `review:critical` / `review:skip` から 1 つを選び、`.coderabbit.config.ts` がラベルで profile と工程別の指示を切り替える。CI `Review Policy` がラベルが 1 つだけ付いていることと設定の型・分岐テストを検査する（[PR 運用ルール](../rules/git-pull-request.md)） |
| テストの有無の検査 | CI `Test Presence` に置き換える。API の operation ごとのテスト（BE の Handler・FE の生成 MSW ハンドラ）が無ければ失敗、変更したソースの同じディレクトリにテストの変更が無ければ警告。正は CI にする（人の PR も検査するため）。Stop hook `require-tests-on-change.sh` は AI の作業時の先行チェックとして残す（判定は CI と同じではない） |
| CI で DB を使う | Backend CI の `check` ジョブで PostgreSQL を起動して migration を当て、DB を使うテストは繋がらなければ skip ではなく失敗にする（[2026-09-29 の ADR](./2026-09-29-presentation-infra-integration-tests.md) で却下した案 E を採用し直す。結合テストと配線のテストが増え、skip のまま通ると何も検証しないため） |
| Go の lint | golangci-lint に `gosec`・`modernize` を追加 |
| ESLint | `tseslint.configs.strict` + `stylistic`、`@eslint-react/eslint-plugin`、`eslint-plugin-react-hooks` の recommended セットに強化。feature 間 import・共有コードから features への import・別ディレクトリへの `../` 相対 import・default export（Story を除く）を機械的に禁止（依存追加: `@eslint-react/eslint-plugin`） |
| Git | `docs/` ブランチと `docs` コミット type を追加 |
| 文書 | [コメントの書き方](../rules/code-comments.md)・[開発の流れ](../rules/development-flow.md)・`AGENTS.md`（Claude Code 以外のエージェントの入口）を追加。`backend/architecture.md` に命名規則と層ごとの規則一覧、`frontend/architecture.md` に層ごとの規則と名前の付け方を追加 |
| 生成物の表示 | `.gitattributes` の `linguist-generated` で、swag の排出物と Orval の生成物を PR の差分で折りたたむ |

## 検討した代替案

| 案 | 概要 | 採否・理由 |
|---|---|---|
| OpenAPI Design First | リポジトリ直下に手書きの YAML を置き、BE（oapi-codegen）・FE（Orval）をそこから生成する | **保留**。契約を実装より先に合意でき、`required`・`enum` を契約に書ける利点はあるが、Handler・入力検証・テストの全面書き換えになる。規則の見直しと混ぜず、別 PR で判断する |
| `internal/<業務領域>/<層>` への再編 | 層ではなく業務領域を第一階層にし、1 つの変更で触るディレクトリを局所化する | **採用**（PR `refactor/area-layout`）。取り込んだ規則を既存の規則より優先する方針に従う。1 つの業務変更が層の数だけのディレクトリに散ることを防ぎ、領域の境界をディレクトリから読めるようにするため。あわせて Handler を閲覧（`public`）と管理（`admin`）に分ける |
| UseCase 名の一括改名 | `ListUsecase` / `GetUsecase` 等を `<動詞><業務対象>Usecase`（`SearchBooksUsecase` / `ReadBookUsecase`）に揃える | **規則は採用。改名は PR `refactor/align-naming-with-rules` で実施**。規則の変更と大規模なリネームは同じ PR にせず、規則の PR の上に積む |
| Node / pnpm / Go の版上げ | ツールチェーンの版を最新に上げる | **見送り**。規則の変更と混ぜない。依存更新の PR で行う |
| CodeRabbit CLI を CI で回す | CI から CLI を実行し、結果を PR にコメントする | **不採用**。GitHub 連携の自動レビューで足りる。レビュー回数の制限を消費するだけ |
| `gocyclo` / ESLint `complexity` の撤去 | 複雑度の上限（10）を外す | **不採用**。分岐の多すぎる関数を読みにくさとして弾く目的は変わらないので維持する |
| Stop hook でのテスト検査を廃止する | CI だけで検査し、hook `require-tests-on-change.sh` を消す | **保留**。AI の作業中に止める先行チェックとしての価値は残る。CI の結果だけで足りると分かったら消す（`.claude/settings.json` の登録解除と同時に行う） |

## 判断基準

- 機械的に検査できる規則は hook ではなく CI に置き、AI だけでなく人の PR も検査する。hook は CI では検知できないもの（ADR・DB 資料の同時更新）に限る
- 規則の変更と大規模なリファクタ（改名・再編・移行）は PR を分ける
- レビューの強度は PR の性質で人が決め、AI は提案に留める
