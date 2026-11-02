# ツールチェーンと依存の版を上げ、脆弱性の検査を CI で必須にする

**日付:** 2026-11-02
**状態:** 採用
**カテゴリ:** library
**参照:** [開発規則の見直し](./2026-11-02-revise-development-rules.md)（版上げを別 PR にした判断）

## 背景・経緯

- Backend CI の `govulncheck` が、Go 1.25.14 の標準ライブラリ（`crypto/x509`・`net/http`）と `golang.org/x/net`・`golang.org/x/text` の、コードから到達する脆弱性で失敗していた
- Frontend CI の `pnpm audit --audit-level high` は、high の脆弱性（`jsdom` 配下の `undici`・`react-router` 等）が残っていたため `continue-on-error` で素通りさせており、検査として機能していなかった
- ESLint の警告が 11 件残り、React 19 で不要になった書き方（`forwardRef`）や依存配列の抑止が混じっていた

## 決定

| 対象 | 版 |
|---|---|
| Go | 1.27.3（`.mise.toml`・`backend/go.mod`） |
| Node | 26.13.0（`.mise.toml`・`frontend/package.json` の `engines`・CI） |
| pnpm | 12.16.1（`.mise.toml`・`packageManager`・`engines`） |
| golangci-lint | v2.15.0（`backend/Makefile`・Backend CI） |
| Go の依存 | echo v4.16.3・gorm v1.31.2・pgx v5.12.0・x/crypto v0.58.0・x/net v0.62.0 ほか最新 |
| フロントの依存 | React 19.3・TanStack Query 5.105・Storybook 10.6・Vite 8.3・Vitest 4.1.12・Orval 8.46 ほか |

- `pnpm audit --audit-level high` の失敗で CI を落とす（`continue-on-error` を外す）
- pnpm 12 はインストール時のスクリプトを既定で止めるため、使わない `esbuild`・`msw` を `pnpm-workspace.yaml` の `allowBuilds` で明示する
- swag は `--parseDependency` を付けると Go 1.27 の標準ライブラリを解析できずに失敗するため外す（DTO は `internal` にしか無く、排出物は変わらない）
- ESLint の警告は 0 件にする。`packages/ui` の `Slot` は子へ props を合成すること自体が役割なので、`cloneElement` 等の規則だけ理由を書いて抑止する

## 検討した代替案

| 案 | 概要 | 採否・理由 |
|---|---|---|
| Go 1.27.3・Node 26・pnpm 12 に上げる（採用） | 規則に揃えた最新の安定版 | 脆弱性の修正版を含み、開発者の環境と CI の版を 1 つに揃えられる |
| Go 1.26.x の最小の修正版に留める | govulncheck が通る最小の版上げ | 不採用。規則で揃える版と開発者の環境を 1 つにするため、修正版だけを拾う中途半端な状態を作らない |
| audit を非ブロッキングのまま残す | 依存の脆弱性を警告だけにする | 不採用。依存更新で high は 0 件にでき、残すと検査が形だけになる |
| `pnpm.overrides` で脆弱な推移的依存を固定する | 直接の依存を上げずに推移的依存だけ上げる | 不要。範囲内の更新（`pnpm update -r`）で high が解消した |

## 判断基準

- ツール・依存の版は、原則として公開から 2 週間以上たったものを選んで固定する。セキュリティ修正版は例外として早めに取り込む
- 版を上げるときは `.mise.toml`・`go.mod`・`frontend/package.json` の `packageManager` / `engines`・CI の版を同じ PR で揃える
