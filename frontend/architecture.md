# フロントエンド方針

## 文書の分担（MECE）

| 文書 | 管轄 |
|---|---|
| [`docs/architecture.md`](../docs/architecture.md) | リポジトリ全体・トップレベル構成・契約の**境界** |
| [`backend/architecture.md`](../backend/architecture.md) | バックエンドのスタック・層・HTTP/OpenAPI |
| **本書** | フロントのスタック・設計パターン・アプリ内ディレクトリ・共有ルール |

本書にバックエンド実装の詳細は書かない。  
OpenAPI は **バックエンドが swag で排出したもの**が入力。フロントは **生成物（型／クライアント）の利用**だけ扱う。排出・HTTP の正は [`backend/architecture.md`](../backend/architecture.md)。  
人が読む方針は **日本語**（AI 入口の英語ルールは `CLAUDE.md` / `.cursor/rules`）。

参照: [bulletproof-react — Project Structure](https://github.com/alan2207/bulletproof-react/blob/master/docs/project-structure.md)  
アプリ分割・共有制限は本リポジトリ方針が優先。ディレクトリの型は bulletproof に寄せる。

## 1. スタックと配信

| 項目 | 決定 |
|---|---|
| フレームワーク | **React + Vite**（**Next.js は使わない**。SSR / RSC は使わない） |
| UI | **デジタル庁 DS**（`@digital-go-jp/tailwind-theme-plugin` + 必要部品を `packages/ui` に取り込み）。**shadcn 不採用** |
| 言語 | **TypeScript 必須**（`.js` / `.jsx` のアプリコードは置かない。`strict` を前提） |
| パッケージマネージャ | **pnpm**（版はピン固定） |
| ツールチェーン | **mise**（Node / pnpm）。**Docker / Dev Container は廃止**（ローカルも CI もホスト上の Node） |
| テスト | **Vitest** + **React Testing Library**（§8） |
| 成果物 | Vite の静的ビルド（`dist/`） |
| ルーティング | クライアント側（React Router）。配信先は SPA フォールバック（`index.html`）を前提 |

API はバックエンド（Go/Echo）へ。フロントは静的アセットのみをホストする。

### 1.1 バージョン固定（mise + pnpm）

メンバーごとの Node / pnpm 差で開けない状態を避ける。

| 層 | 固定手段 | ファイル |
|---|---|---|
| Node / pnpm | **mise** | リポジトリ直下 `.mise.toml` |
| pnpm 本体（再確認） | `packageManager` + `engines` + `engine-strict` | `frontend/package.json` / `frontend/.npmrc` |
| ライブラリ | `pnpm-lock.yaml` **1本**（workspace ルート） | `frontend/pnpm-lock.yaml` |

- **pnpm workspace は lock 1本**（`frontend/` ルート）
- 版上げは `.mise.toml` と `frontend/package.json` の `packageManager` / `engines` を **同じ PR** で更新する
- セットアップ: `mise install` → `cd frontend && pnpm install`（手順は [README.md](./README.md)）

## 2. アプリと共有

- **Web アプリ**: `frontend/web`（書籍・著者の管理画面。React + Vite + TS）
- **共有してよいもの**: `packages/ui`（デジタル庁 DS のテーマ＋取り込み部品）、OpenAPI 生成型、純関数（金額・日付）
- **共有しないもの**: ドメイン入りの複合コンポーネントを何画面からも呼ぶこと
- 同一アプリ内でも、似ているだけの UI はコピーを許容。3回同じになったら抽出を検討
- アプリ横断で共有するものだけ `frontend/packages/` に置く（中身は上記「共有してよいもの」に限定）
- 各 Web アプリ・packages は **pnpm workspace（lock 1本）** で管理する
- feature 間 import 禁止は最初 **文書のみ**。痛くなったら ESLint `import/no-restricted-paths`

### 2.0 端末・画面幅の前提

- **PC を主**とし、一覧・作業効率を優先する。スマホでも主要操作が破綻しないレスポンシブにする
- Storybook / レビュー時の確認幅: 広幅を主、狭幅でも崩れないこと

### 2.1 デジタル庁 DS の usage 参照（いつ見るか）

部品の使い方・a11y・禁止事項の正は公式: https://design.digital.go.jp/dads/  
（例: [Accordion usage](https://design.digital.go.jp/dads/components/accordion/usage/)）

| タイミング | usage を見る？ |
|---|---|
| **初めて**その部品を `packages/ui` に取り込む | **見る** |
| 既存部品の配置・文言だけの画面実装 | **見ない**（`packages/ui` と Storybook が正） |
| 挙動や a11y で迷った／レビューで指摘された | **見る** |
| 毎回すべてのコンポーネントの usage | **しない**（過剰） |

**守らせ方**

1. 本表を正とする（本書 + `.cursor/rules`）
2. `packages/ui` に部品を追加する PR は、本文に **該当 usage URL を1行**書く
3. 画面だけの PR では usage 確認を求めない
4. AI／実装者は「新規取り込み・迷ったとき」以外に usage 全文を読まない

## 3. 設計パターン

**Container/Presenter の明示2層（専用フォルダ・命名）は採用しない。**

| 役割 | 置き場 | 責務 |
|---|---|---|
| ロジック | `hooks/`（feature 内を基本） | 取得・変換・イベント・画面状態の組み立て |
| 表示 | コンポーネント | **props で受け取り描画に徹する**（Presentational に保つ） |
| 画面のつなぎ | `app/`（routes / provider） | feature の hook + 表示コンポーネントを合成するだけ |

- コンポーネント内にデータ取得や複雑な分岐を溜めない。必要なら hook へ上げる
- 「smart / dumb」の分離は **責務で行い、ディレクトリ名では表現しない**

### 3.1 Storybook（Presentational の状態カタログ）

**入れる。ただし画面全体は対象外。**  
目的は人間・AI が **表示のバリエーションを把握しやすくする**こと。

| やる | やらない |
|---|---|
| props で完結する Presentational のストーリー | ページ丸ごと（Router + Query + 認証込み） |
| 状態カタログ: `Default` / `Loading` / `Empty` / `Error` / `Disabled` など | 本番 Provider や API をそのまま載せること |
| 短い description（何の UI か・いつ使うか・ドメイン用語） | store / `useQuery` 直結コンポーネントの無理なストーリー化 |

置き場:

```
features/<feature>/components/Foo.tsx
features/<feature>/components/Foo.stories.tsx
components/ui/Button.stories.tsx   # アプリ内 primitive
```

- Storybook は **アプリごと**（アプリ横断の複合 UI カタログは作らない）
- decorator はテーマ等の最小限。Query が必要ならそのストーリー用の mock QueryClient のみ

### 3.2 CI / ハーネス

PR で必須:

- `typecheck` / `lint`（**ESLint**）/ `format:check`（**Prettier**）/ `test` / `build` / `build-storybook`（**Storybook はビルド成功のみ**。視覚回帰はしない）
- ESLint（flat config）と Prettier の設定は workspace ルート（`frontend/eslint.config.mjs` / `.prettierrc`）に1つ

当面やらない:

- **Playwright / E2E ランナー導入**（基幹導線ができてから。方針は §8）
- **画像比較**（必要画面が出てから Playwright screenshot または Chromatic 等）
- **カバレッジ閾値**（設けない）

## 4. 状態・データの使い分け

| 種類 | 手段 | 用途 |
|---|---|---|
| サーバ状態 | **TanStack Query** | API 取得・キャッシュ・再取得・ミューテーション |
| クライアント UI 状態 | **Zustand** | モーダル開閉、ウィザード途中、横断トースト等（サーバと無関係なもの） |
| 横断的注入 | **Provider（React Context）** | テーマ、認証セッション、i18n、QueryClient などアプリ全体の配線 |

- サーバ由来のデータを Zustand に二重保管しない（Query を正とする）
- Context にビジネスデータや頻繁に変わる値を載せない（再レンダー範囲が広がる）

## 5. OpenAPI 生成物の利用（フロント側）

契約・分割・再生成フローの正は [`docs/architecture.md`](../docs/architecture.md) §4。

```bash
# リポジトリ frontend/ で（backend/api/docs/swagger.yaml が入力）
pnpm gen:api
```

生成は **Orval**（設定 `web/orval.config.ts`。swag の Swagger 2.0 を Orval が読み込み時に OpenAPI 3 へ変換）。

| 生成物 | 場所 |
|---|---|
| TanStack Query フック・リクエスト関数 | `web/src/api/generated/api.ts` |
| 型（DTO） | `web/src/api/generated/api.schemas.ts` |
| MSW ハンドラ（テスト用） | `web/src/api/generated/api.msw.ts`（faker のモック値は `api.faker.ts`） |
| fetch の共通処理（手書き） | `web/src/api/mutator.ts`（`VITE_API_BASE_URL` 付与・`ErrorResponse` → `ApiError`） |

- 生成物（`web/src/api/generated/`）は手編集しない。契約を更新して再生成する。ESLint / Prettier の対象外
- 画面は生成フックを feature の hook から呼ぶ（Query フックを手書きしない）
- テストの既定ハンドラは生成 MSW ハンドラ（paths が空の間は Orval が生成しないため空配列）。具体値を検証するテストは `server.use(get…MockHandler(fixture))` で上書きする
- ローカル: Vite が `/api` と `/health` を `localhost:8080` にプロキシ。本番ビルドは `VITE_API_BASE_URL` を指定
- CI: `pnpm gen:api:check`（再生成して `git diff --exit-code` と未追跡ファイルの有無を確認）

## 6. アプリ内ディレクトリ（bulletproof 準拠）

各アプリ（例: `frontend/web/`）の `src/` は次を基本とする。不要なフォルダは作らない。

```
src/
  app/                 # router / provider / アプリ入口（Vite + クライアントルーティング）
  api/                 # Orval 生成物（generated/）と mutator（横断の HTTP）
  assets/
  components/          # アプリ内共有 UI（primitive・レイアウト。ドメイン複合は非推奨）
  config/
  features/            # 機能モジュール（ここに大半を置く）
  hooks/               # アプリ横断の薄い hooks のみ
  lib/                 # 事前設定済みライブラリ（queryClient 等）
  stores/              # Zustand（アプリ横断）
  testing/             # 共有テスト用具・MSW。E2E など横断テスト（単体はコロケーション）
  types/               # 共有型（画面側のドメイン型。生成 API 型は api/generated/）
  utils/
```

feature の中身（必要なものだけ）:

```
features/<feature>/
  api/          # 生成フックで足りないリクエスト組み立て（稀。Query フックは Orval 生成を使う）
  components/   # feature スコープの表示コンポーネント（props 中心）
  hooks/        # feature 固有ロジック
  stores/       # feature 限定のクライアント状態（稀）
  types/
  utils/
  *.test.ts     # 単体・画面テストはソース横（コロケーション）
```

## 7. 依存の向き

bulletproof と同様、**一方向**にする。

```
shared（components / hooks / lib / types / utils）
    ↑
features（feature 同士は import しない。合成は app で行う）
    ↑
app（routes / provider）
```

- feature 間の横断 import は禁止。必要なら共通化を `packages/` か shared へ上げるか、`app` で合成する
- barrel（`index.ts` の再エクスポート祭り）は必須にしない（バンドラの tree-shaking を優先し、直接 import）

## 8. テスト（Vitest + React Testing Library）

目的は件数・カバレッジではなく、**不具合を早く見つけ安心して変更できること**。  
**カバレッジ閾値は設けない。**  
ランナー **Vitest** / 画面 **React Testing Library**（ユーザー視点）/ API 境界は **MSW**。  
Storybook（§3.1）は表示カタログ。テストと役割を混ぜない。

**置き場**

| 種類 | 場所 |
|---|---|
| 単体・画面（`*.test.ts` / `*.test.tsx`） | **ソース横（コロケーション）** |
| MSW handlers・テスト用具・E2E 等の横断 | `src/testing/` |

```
計算・判定・hooks/store → 単体を厚く
フォーム・重要画面     → 操作と表示の画面テスト（必要なら MSW）
書籍・著者の登録更新など基幹導線 → E2E は少数（ランナー導入後）
```

**E2E 環境（導入後）**: CI は **mock / MSW 寄り**。stg 向けは **少数の煙**のみ。

| 対象 | 方針 |
|---|---|
| 計算・バリデーション・表示条件・hooks | 単体を厚く（境界値・不正値・組み合わせ） |
| フォーム・重要画面 | 操作と表示。重要なら MSW で API 境界まで |
| 単純表示（文言・props 透過・アイコン） | 個別テストしない（上位で触れれば十分） |
| 書籍・著者の登録・更新等 | E2E は将来少数（**当面 Playwright 未導入**）。細部は単体/画面へ |
| CSS | 必要な画面だけ画像比較（**当面未導入**） |

**モックする**: HTTP / Analytics・外部認証 / 日時など外部依存。  
**モックしない**: 自作の子コンポーネント（実物で組み立てる）。  
重要画面の状態は必要分だけ: ローディング / 成功 / 0件 / 入力エラー / API エラー / 操作成功。  
確認するのは表示・入力・押下可否・エラー・操作後の変化・API が正しい内容で呼ばれたか。内部 state / private は見ない。  
Snapshot は主力にしない（明示 assert）。小さく安定した部品に限定可。

**どれにするか**: 入出力だけで足りる条件判定 → 単体。操作・表示切替・フォーム → 画面。止まると痛い複数画面＋接続 → E2E（導入後）。

### 鉄則

1. 業務判断や計算は UI から分離して単体テストする  
2. コンポーネントはユーザー操作と表示結果をテストする  
3. private な状態や内部メソッドを検証しない  
4. 自作の子コンポーネントを過剰にモックしない  
5. 重要画面は API 境界をモックした統合テストにする  
6. ローディング・空・成功・エラーの主要状態を確認する  
7. 単純表示コンポーネントはテストしすぎない  
8. 巨大 Snapshot を主力にしない  
9. E2E は業務継続に直結する導線だけ  
10. CSS 崩れは必要な画面だけ画像比較で検証する  
