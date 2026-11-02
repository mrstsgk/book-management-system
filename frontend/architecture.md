# フロントエンド方針

## 文書の分担（MECE）

| 文書 | 管轄 |
|---|---|
| [`docs/architecture.md`](../docs/architecture.md) | リポジトリ全体・トップレベル構成・契約の**境界** |
| [`backend/architecture.md`](../backend/architecture.md) | バックエンドのスタック・層・HTTP/OpenAPI |
| **本書** | フロントのスタック・設計パターン・アプリ内ディレクトリ・共有ルール |

本書にバックエンド実装の詳細は書かない。  
OpenAPI はリポジトリ直下 `openapi/` の YAML（API 契約の正本）をバンドルした `openapi/dist/web.yaml` が入力。フロントは **生成物（型／クライアント／zod）の利用**だけ扱う。契約の正は [`docs/architecture.md`](../docs/architecture.md) §4。  
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

- **Web アプリ**: `frontend/web`（読んだ本の紹介画面と、自分用の登録・編集画面。React + Vite + TS）
- **共有してよいもの**: `packages/ui`（デジタル庁 DS のテーマ＋取り込み部品）、OpenAPI 生成型、純関数（金額・日付）
- **共有しないもの**: ドメイン入りの複合コンポーネントを何画面からも呼ぶこと
- 同一アプリ内でも、似ているだけの UI はコピーを許容。3回同じになったら抽出を検討
- アプリ横断で共有するものだけ `frontend/packages/` に置く（中身は上記「共有してよいもの」に限定）
- 各 Web アプリ・packages は **pnpm workspace（lock 1本）** で管理する
- feature 間 import の禁止は ESLint で機械的に検査する（§7）

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

### 2.2 入力チェックのメッセージ

入力チェックのメッセージは、設計モック・実装のどこで書くときも次の 4 つを守る。

1. **ユーザーを責めない**: 「〇〇が間違っています」「入力ミスです」のような否定の言葉や、「エラー」「失敗」のような強い言葉を使わない
2. **具体的な形式や条件を示す**: 「半角数字で入力してください」「5000 文字以内で入力してください」「ハイフンは不要です」のように、次に何をすればよいかを書く
3. **専門用語を使わない**: 「Null」「Undefined」「トークンが無効です」のようなシステムの言葉を使わず、誰でもわかる言葉にする
4. **入力欄のすぐ近くに表示する**: 該当する入力欄のすぐ下か横に出す。画面の上部にまとめるだけにしない。入力欄に対応しないもの（API の 400 で `errors` を含まないもの、通信失敗）は画面の上部に出す

API の `ErrorResponse` の `message` は開発者向けなので、そのまま画面に出さない。文言は FE が `errors[].field` / `rule` から引く。文言の正は次のとおり。

| 文言 | 正 |
|---|---|
| 制約の種類ごとの共通の文言（`rule` が `required`・`maxLength`・`minimum`・`maxItems`・`format` など OpenAPI のキーワード） | FE の対応表 `ERROR_MESSAGES`（`web/src/utils/errorMessages.ts`。キーは `<field>.<rule>`。本の登録・編集・タグ・ログインで共有するため shared に置く） |
| 相関ルール・業務ルールに固有の文言（`rule` が `isbnDuplicated` などの id、`code` が `versionConflict` などのコード） | 概要設計 `docs/design/<機能>.md` のメッセージ一覧（MSG-xx）。同じ対応表に写し、`getErrorMessage(key, params)` で引く（`params` の値を `{キー}` に差し込む。値が欠けたら呼び出し側の既定の文言にする） |
| 5xx・通信失敗・`errors` の無い 400 | FE 共通の文言（画面上部） |

上限のある項目の文字数は、BE と同じく前後の空白を除いて Unicode のコードポイントで数える（共通の `utils/codePointLength.ts`。生成 zod の `.max()` も同じ数え方。`String.length` と HTML の `maxlength` は UTF-16 単位で絵文字を 2 文字と数えるため使わない。`maxlength` は IME の変換途中でも文字を切る）。上限の値は生成 zod の定数（`registerBookBodySummaryMax` など）を使い、FE で書かない。上限を超えたら送信を止め、入力欄のすぐ下で伝える（実例: 一覧の検索キーワード `features/books-list/utils/searchParams.ts`）。

`errors[]` は順序に依存せず全件を表示し、`field` に対応する入力欄のすぐ下に出す（[`docs/architecture.md`](../docs/architecture.md) §4「バリデーションの全件返却」）。

## 3. 設計パターン

**Container/Presenter の明示2層（専用フォルダ・命名）は採用しない。**

| 役割 | 置き場 | 責務 |
|---|---|---|
| ロジック | `hooks/`（feature 内を基本） | 取得・変換・イベント・画面状態の組み立て |
| 表示 | コンポーネント | **props で受け取り描画に徹する**（Presentational に保つ） |
| 画面のつなぎ | `app/`（routes / provider） | feature の hook + 表示コンポーネントを合成するだけ |

- コンポーネント内にデータ取得や複雑な分岐を溜めない。必要なら hook へ上げる
- 「smart / dumb」の分離は **責務で行い、ディレクトリ名では表現しない**

**View State**: 主要な画面状態は明示する。複数の boolean の組み合わせで 1 つの画面状態を表し始めたら、discriminated union にまとめる。

```ts
type BooksListViewState =
  | { type: 'loading' }
  | { type: 'empty' }
  | { type: 'error'; message: string }
  | { type: 'ready'; items: BookListItem[]; hasMore: boolean }
```

**API Model ≠ UI Model ≠ Props**: OpenAPI 生成型は API 境界の契約であり、UI の型とは限らない。API の構造と画面が必要な構造が違うときだけ、feature の hook か mapper で UI 向けの形に変換する。単純な受け渡しのためだけに UI Model や mapper を機械的に作らない。

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

- **E2E を PR 必須（CI）に含めること**（Playwright は導入済みだがローカル実行のみ。方針は §8）
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
# リポジトリ frontend/ で（../openapi/dist/web.yaml が入力。YAML を変えたら先に make -C ../backend openapi）
pnpm gen:api
```

生成は **Orval**（設定 `web/orval.config.ts`）。hook・関数の名前は operationId から付く（`searchBooks` → `useSearchBooks`・`getSearchBooksMockHandler`）。

| 生成物 | 場所 |
|---|---|
| TanStack Query フック・リクエスト関数 | `web/src/api/generated/api.ts` |
| 型（DTO） | `web/src/api/generated/api.schemas.ts` |
| MSW ハンドラ（テスト用） | `web/src/api/generated/api.msw.ts`（faker のモック値は `api.faker.ts`） |
| zod スキーマ・制約の定数（フォームの入力チェック用） | `web/src/api/generated/zod/api.ts` |
| fetch の共通処理（手書き） | `web/src/api/mutator.ts`（`VITE_API_BASE_URL` 付与・`ErrorResponse` → `ApiError`（`status`・`fieldErrors`・`code`・`params`）） |

- 生成物（`web/src/api/generated/`）は手編集しない。契約を更新して再生成する。ESLint / Prettier の対象外。`.gitattributes` の `linguist-generated` で PR の差分では折りたたまれる
- 画面は生成フックを feature の hook から呼ぶ（Query フックを手書きしない）
- テストの既定ハンドラは生成 MSW ハンドラ。具体値を検証するテストは `server.use(get…MockHandler(fixture))` で上書きする。**生成 API の差し替えは MSW で行い、`vi.mock()` しない**（成功は生成ハンドラ、4xx やリクエスト内容の検証は同じメソッド・パスの手書き `http.put('*/api/admin/tags/:id', …)`。API 契約の Mock として扱うため。CI `Test Presence` もこの形を検出する）
- 契約で必須になった項目を満たすテスト用データは、乱数の faker ではなく `src/testing/fixtures.ts` の `createXxxFixture` で作る（書影の有無などを確かめるテストが不安定にならないように）
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

次は ESLint（`frontend/eslint.config.mjs`）で機械的に検査する。feature を足したら設定の一覧にも足す。

| 検査 | 内容 |
|---|---|
| feature 同士の import 禁止 | `features/<a>` から `@/features/<b>` を import しない |
| 共有コードから features への import 禁止 | `api` / `components` / `hooks` / `lib` / `testing` / `types` / `utils` は `@/features/**` を import しない（依存は shared ← features ← app の向きだけ） |
| 別ディレクトリへの `../` 相対 import 禁止 | 別ディレクトリは `@/` から書く。同じディレクトリと子ディレクトリは `./` |
| default export 禁止 | named export にする（Storybook の `meta` だけ例外） |

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
読んだ本の登録・更新など基幹導線 → E2E は少数（Playwright、ローカル実行のみ）
```

**E2E（Playwright、導入済み）**: 本のライフサイクル・タグのライフサイクルの2導線のみ、実バックエンド・実 DB・実フロントエンドサーバーに対して**ローカルでのみ**実行する。**CI には載せない**（実DB接続が要り重いため。判断は [ADR](../docs/adr/2026-09-30-introduce-playwright-e2e.md)）。置き場は `web/src/testing/e2e/`。

| 対象 | 方針 |
|---|---|
| 計算・バリデーション・表示条件・hooks | 単体を厚く（境界値・不正値・組み合わせ） |
| フォーム・重要画面 | 操作と表示。重要なら MSW で API 境界まで |
| 単純表示（文言・props 透過・アイコン） | 個別テストしない（上位で触れれば十分） |
| 読んだ本の登録・更新等 | E2E は少数（本のライフサイクル・タグのライフサイクルの2つ）。分岐網羅・境界値は単体/画面へ |
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

## 9. 層ごとの規則と名前の付け方

層（ディレクトリ）ごとに散らばっている規則の索引。正は各節（設計は本書、テストの書き方は [テスト規約](../docs/rules/testing.md)、コメントは [コメントの書き方](../docs/rules/code-comments.md)）。「実例」は現在の `web` のコード。

| 層 | 置き場・ファイル名 | 命名・書き方 | テスト |
|---|---|---|---|
| **app** | `src/app/`。`App.tsx`・`provider.tsx`・`router.tsx`・`AdminGuard.tsx`・`routes/XxxRoute.tsx`（実例: `routes/BooksListRoute.tsx`） | Router / Provider / feature の合成だけを行う（§3）。route は feature の hook と表示コンポーネントをつなぐ | 合成だけなら個別のテストは書かない。認証による遷移の振り分けなど挙動を持つもの（`AdminGuard`・`router`）は同名の `*.test.tsx` |
| **api** | `src/api/`。`generated/`（Orval の生成物。手編集しない）と `mutator.ts` | 全リクエストは `apiMutator` を通り、失敗は `ApiError` に正規化する。hook 名・関数名は operationId から生成（`useSearchBooks`）。フォームの上限は `generated/zod/` の定数 | `mutator.test.ts`: MSW で成功・204・HTTP エラー・通信失敗・不正な JSON を検証。生成物はテストしない |
| **features（画面）** | `src/features/<feature>/components/`。`XxxView.tsx`・`XxxForm.tsx`（実例: `books-list/components/BooksListView.tsx`、`admin-book-editor/components/BookForm.tsx`） | PascalCase のファイル・同名の named export（`export function BooksListView`）。API 取得・大きな変換・複雑な状態遷移は hook に出す | 同名の `*.test.tsx`。MSW の生成ハンドラを `server.use(...)` で上書きし、主要状態（読み込み中 / 0 件 / 成功 / API エラー）ごとに見える内容と押せるものを検証 |
| **features（hook）** | `src/features/<feature>/hooks/useXxx.ts`（実例: `useBooksList.ts`、`useBookEditor.ts`、`useLoginForm.ts`） | hook が提供するもので名付ける（下の「名前の付け方」）。API の構造と画面が必要な構造が違うときだけ生成 hook を包む | ロジックがあれば直接テストする（`useBooksList.test.tsx`） |
| **features（utils・型）** | `src/features/<feature>/utils/`（実例: `paging.ts`、`validateBookForm.ts`、`barRatio.ts`）、`types.ts`（実例: `admin-book-editor/types.ts`） | 関数名と同じ camelCase のファイル名（1 語なら小文字）。型は PascalCase、値の対応表は UPPER_SNAKE（`PAGE_SIZE`） | 計算・判定は直接呼んで分岐・境界値を検証。型だけならテスト不要 |
| **components（共有）** | `src/components/{layouts,pages,states,ui,form}/`。`layouts/` は枠（`Layout`・`Header`・`Footer`・`AdminLayout`）、`pages/` は feature に属さない画面（`NotFoundPage`）、`states/` は読み込み中・0 件・エラーの表示、`ui/` は小さな部品（`RatingStars`・`ConfirmDialog`）、`form/` はフォームの部品（`FormField`・`ErrorSummary`） | 業務の意味を持つ複合 UI は置かない（§2）。DS の部品は `packages/ui` | ロジック・操作があれば同名の `*.test.tsx`。Story は `*.stories.tsx`（§3.1） |
| **hooks / lib / testing** | `src/hooks/`（アプリ横断の薄い hook。実例: `useLoginRedirect.ts`）、`src/lib/`（ライブラリの設定。実例: `query-client.ts` の `createQueryClient`）、`src/testing/`（`server.ts`・`handlers.ts`・`render.tsx`・`setup.ts`・`e2e/`） | lib は `createXxx` の生成関数。`testing/` は未定義のリクエストをエラーにする | hook はロジックがあれば同名のテスト。`testing/` はテストしない |
| **packages/ui** | `frontend/packages/ui/src/`。部品ごとに PascalCase のファイル、`index.ts` から export | デジタル庁 DS の部品だけ。新規追加時は usage を読み、PR に URL を書く（§2.1） | ロジックがあれば同名の `*.test.tsx` |
| **全層共通** | 別ディレクトリの import は `@/` から、同じディレクトリと子は `./`（§7）。barrel は必須にしない | named export（default export は Storybook の `meta` だけ）。feature 同士は直接 import しない（§7） | テストは対象と同名で同じディレクトリ。テスト名は日本語で期待する振る舞い。入力違いは `it.each`。Snapshot・Coverage の閾値は使わない（§8） |

### 名前の付け方

命名の規則はこの節が正。名前は目的駆動で、存在や技術的な手段ではなく何のためのものかから付ける（[`backend/architecture.md`](../backend/architecture.md) §2「設計原則」と同じ考え方）。

- Component は何を表示し何を操作できるか、hook は何を提供するか、関数は何をするか、型は何を表すかで名付ける
- 業務の用語（要件定義と同じ言葉: 本・分野タグ・一言まとめ・感想・評価）を揃え、違う概念には違う名前を使う
- `Manager` / `Helper` / `Common` / `Data` / `Processor` / `Util` のような広すぎる名前は、具体的な責務で言い換えられないか見直す（機械的に禁止はしない）
- 短さより、語を足して意味が増えるかで判断する。名前を揃えるためだけに目的の違うものをまとめたり、区別の要らないものを分けたりしない

**表記**

| 対象 | 規則 | 例 |
|---|---|---|
| Component | PascalCase。役割と表示対象を表す（`View` / `Form` / `Dialog` などは実際の役割と合わせる）。`XxxComponent` / `XxxUI` / `XxxManager` にしない | `BooksListView`、`ConfirmDialog` |
| 画面（route） | `XxxRoute`（`app/routes/`）。feature に属さない画面は `XxxPage` | `BookDetailRoute`、`NotFoundPage` |
| Component の宣言 | `function` 宣言にし、ファイル名・export 名と揃える。コールバックや短い内部の関数はアロー関数でよい | `export function BookForm(...)` |
| Props の型 | `XxxProps` | `BooksListViewProps` |
| 型 | PascalCase。`I` の接頭辞を付けない。オブジェクトの型は `interface`、union・交差型・`Record` など `type` でしか書けないものは `type`（ESLint の `consistent-type-definitions` で検査） | `BookFormValues` |
| 関数・変数・props | camelCase。関数は動詞で始め、目的を表す（計算は `calculateXxx`、表示の形に整えるのは `formatXxx`、選ぶ・解釈するのは `resolveXxx` / `parseXxx`、検証は `validateXxx`） | `parseBookId`、`validateBookForm` |
| 真偽値 | 意味で `is`（状態）/ `has`（持っている・ある）/ `can`（できる）を使い分ける | `isLoading`、`hasMore`、`canSubmit` |
| イベントの props | `onXxx` | `onSubmit`、`onClose` |
| Component の中のイベント処理 | `handleXxx` | `handleSubmit` |
| state の setter | `setXxx` | `setValues` |
| 値の対応表・上限の定数 | UPPER_SNAKE | `PAGE_SIZE`、`ERROR_MESSAGES` |
| ほかの定数 | 意味に合わせて付ける（すべてを UPPER_SNAKE にはしない） | `emptyBookFormValues` |

**hook**

- hook が外に提供する状態・情報・操作で名付ける。画面名・API の関数名・実装の技術から機械的に付けない（`useBooksList` / `useBookEditor` / `useLoginForm`）
- `Query` / `Mutation` の接尾辞は、区別が要るときだけ付ける
- 生成 hook は operationId から付いた名前（`useSearchBooks`）のまま使い、名前を変えるためだけの中間 hook を作らない（§5）。feature の hook は画面の目的で名付け、生成 hook の名前に揃えようとしない。名前がぶつかるとき（生成の `useLogin` と画面のフォーム）は、feature 側を目的で付け直す（`useLoginForm`・`useLogoutAction`）

**API Model → UI Model の変換（mapper）**

- 変換先が 1 つで明らかなら `toXxx`（`toBookFormValues`）、変換元・変換先が曖昧なら `mapXxxToYyy`。`convertData` / `adapt` のような抽象的な名前にしない
- UI Model は画面での目的・意味で名付け、`Model` / `View` の接尾辞は必須にしない。API Model と同じ形で変換が要らなければ mapper を作らない（§3）

**Zustand の store**

- hook の名前は `useXxxStore`、ファイルも同じ名前。状態の型は `XxxState`、操作は具体的な動詞
- 置き場は最も狭いスコープから選ぶ（§4）。store の名前を決めるためだけに store を作らない（現在は store なし）

**feature**

- `src/features/<feature>/` は業務の領域か、利用者に提供する機能で名付ける（実例: `books-list` / `book-detail` / `tag-counts` / `admin-book-editor` / `admin-login` / `admin-tags`）。管理画面側は `admin-` を前に付けて公開画面と区別する
- `common` / `misc` のように、何が属するかの基準が曖昧な名前を作らない

**ファイル**

| 種類 | 規則 | 例 |
|---|---|---|
| Component・画面・Story・テスト | export する Component の名前と同じ PascalCase | `BookForm.tsx`・`BookForm.test.tsx`・`BookForm.stories.tsx` |
| hook・store | hook の名前と同じ camelCase | `useBooksList.ts` |
| utils の関数 | 主となる export の関数名と同じ camelCase（本リポジトリの実例に揃える。kebab-case にしない） | `validateBookForm.ts`・`parseBookId.ts`・`barRatio.ts` |
| 設定・複数の export を持つ `.ts` | kebab-case か 1 語の小文字 | `query-client.ts`・`paging.ts`・`types.ts` |

JSX を含むファイルは `.tsx`、含まないファイルは `.ts`。

**テスト用のデータ・その他**

- fixture を作る関数は `createXxxFixture`。エラーのコードから文言を引く対応表は `ERROR_MESSAGES`、引く関数は `getErrorMessage` を基本にする。API の通信エラーは `ApiError` に正規化し、Component ごとにエラーの型を増やさない

## 10. 未決の規則（必要になったときに決める）

まだ実例が無い、または現在のコードで書き方が揃っていないもの。該当する実装が初めて出る PR で決め、決めたら §9 と各節に反映する。推測で先に決めない。

| 対象 | 決めること |
|---|---|
| Scenario / Fixture | 画面の状態（0 件・多件・API エラー等）を表すテストデータの置き場と、どこまで再利用するか（現在はテストごとにその場で書いている） |
| ルーティング | path の定義の置き場（現在は `router.tsx` と各 Component に文字列で書いている）と、画面と route の対応の書き方（案: `bookPaths.detail(id)` のように目的を表す名前でまとめる） |
| hook のテスト | `renderHook` で hook を直接テストする書き方と、生成 hook の差し替え方（現在は MSW で差し替えている） |
| Storybook | Story の置き場（`packages/ui` の `Button` の Story がアプリ側 `components/ui/Button.stories.tsx` にある） |
