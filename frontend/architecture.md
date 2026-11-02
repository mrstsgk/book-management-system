# フロントエンド方針

## 文書の分担

| 文書 | 管轄 |
|---|---|
| [`docs/architecture.md`](../docs/architecture.md) | リポジトリ全体の方針・チャネル／契約の**境界** |
| [`backend/architecture.md`](../backend/architecture.md) | バックエンドのスタック・層・HTTP（OpenAPI からの BE 生成） |
| **本書** | フロントのスタック・設計パターン・アプリ内ディレクトリ・共有ルール |

本書にバックエンド実装の詳細は書かない。

OpenAPI はリポジトリ直下 `openapi/` の YAML（API 契約の正本。[`docs/architecture.md`](../docs/architecture.md) §4）を入力とし、本書では生成物の利用のみ扱う。

人が読む方針は日本語とし、AI向けルールは `CLAUDE.md` / `AGENTS.md` に置く。

参照:

- [bulletproof-react — Project Structure](https://github.com/alan2207/bulletproof-react/blob/master/docs/project-structure.md)
- Apple Human Interface Guidelines（public）
- デジタル庁 Design System（admin）

リポジトリ固有の共有・依存ルールを優先する。

---

## 1. スタックと配信

| 項目 | 決定 |
|---|---|
| フレームワーク | React + Vite |
| 言語 | TypeScript (`strict`) |
| UI | チャネル別: public は Apple HIG（iOS）、admin はデジタル庁 Design System（§4） |
| パッケージマネージャ | pnpm（版は `.mise.toml` と `frontend/package.json` の `packageManager` / `engines` で固定。lock は `frontend/pnpm-lock.yaml` の 1 本） |
| ツールチェーン | mise |
| API生成 | Orval |
| Server State | TanStack Query |
| Client State | React local state / Zustand |
| テスト | Vitest + React Testing Library + MSW |
| 成果物 | Vite static build (`dist/`) |
| 配信 | しない（デプロイは要求の範囲外。ローカルの開発サーバーで動かす） |
| ルーティング | クライアントルーティング（React Router） |

Next.js / SSR / RSC は使用しない。

Docker / Dev Container は使用せず、ローカル・CIともにホスト上のNodeを利用する。

---

## 2. チャネルと共有

Webアプリ:

- public（`frontend/public`。読んだ本の紹介。誰でも見られる）
- admin（`frontend/admin`。自分用のログイン・本と分野タグの登録・編集）

Web横断で共有してよいもの:

- `packages/ui`（デジタル庁 DS。拠り所が DS のチャネル＝admin で使う。§4）
- OpenAPI生成型
- 日付などの純関数
- 技術的な共通処理

ドメイン意味を持つ複合UIは原則共有しない。

似ているだけのUIはコピーを許容し、同じ理由で変更されることが明確になってから共通化する。

feature間の直接importは禁止する。

---

## 3. 端末・画面幅

### public

スマホを主対象とする。

- モバイルファースト
- 1カラム
- 親指操作
- 短い画面単位
- 巨大一覧・多カラム表を避ける

PC では専用レイアウトを作ることを目的にしないが、致命的な崩れや操作不能は避ける。

### admin

PCのみを対象とする。スマホ・タブレット向けの表示は作らない。

PCでは一覧性・作業効率を優先する。

### iPhone 向けの共通事項（public）

Apple HIG（§4）に沿って、iPhone での操作感に合わせる。

- タップできる領域は 44×44pt 以上を確保する
- ノッチ・ホームインジケータに隠れないよう、セーフエリア（`env(safe-area-inset-*)`）を考慮する
- 画面の移動・戻る・モーダル（シート）・タブなど、iOS で見慣れた画面構成と操作に寄せる
- 文字はシステムフォントを基本とし、利用者の文字サイズ設定で崩れないようにする

---

## 4. デザインの拠り所（チャネル別）

チャネルごとに拠り所を分ける。拠り所は主な利用者と端末で決める（public はスマホでの操作感、admin は PC での作業の一貫性とアクセシビリティを重視する）。共通部品の共有は、拠り所が同じチャネルの間に限る。

| チャネル | 拠り所 | 参照 |
|---|---|---|
| public | Apple Human Interface Guidelines と iOS 向けのデザイン | [HIG](https://developer.apple.com/jp/design/human-interface-guidelines)・[Designing for iOS](https://developer.apple.com/jp/design/human-interface-guidelines/designing-for-ios) |
| admin | デジタル庁デザインシステム | [方針](https://www.digital.go.jp/policies/servicedesign/designsystem)・[DADS](https://design.digital.go.jp/dads/) |

アクセシビリティ（ラベル、色だけで伝えない、十分なタップ領域）はどのチャネルでも守る。

### 入力チェックのメッセージ（全チャネル）

入力チェックのメッセージは、モック・概要設計のメッセージ一覧・実装のどこで書くときも次の 4 つを守る。

1. **ユーザーを責めない**: 「〇〇が間違っています」「入力ミスです」のような否定の言葉や、「エラー」「失敗」のような強い言葉を使わない
2. **具体的な形式や条件を示す**: 「半角数字で入力してください」「5000 文字以内で入力してください」「ハイフンは不要です」のように、次に何をすればよいかを書く
3. **専門用語を使わない**: 「Null」「Undefined」「互換性がありません」「トークンが無効です」のようなシステムの言葉を使わず、誰でもわかる言葉にする
4. **入力欄のすぐ近くに表示する**: 該当する入力欄のすぐ下か横に出す。画面の上部にまとめるだけにしない（入力欄に対応しないもの、たとえば API の 400 で `errors` を含まないものは、画面の上部に出す）。入力欄に対応するエラーがあるときは、画面の上部に件数付きの短い案内（「入力内容を確認してください（N件）」）だけを出して個別の文言は繰り返さず、エラーのある最初の入力欄にフォーカスを移す（長いフォームで画面の外の入力欄のエラーに気づけないことと、スクリーンリーダーで保存できなかったことが伝わらないことを防ぐため）

zod と API が返す既定のメッセージ（英語やシステムの言葉）は、そのまま画面に出さない。API の `ErrorResponse` の `message` も開発者向けなので出さない。文言は FE がコードから引く（コードの規則は [`docs/architecture.md`](../docs/architecture.md) §4）。

| コード | 文言の正 |
|---|---|
| 制約の種類（zod の検証結果、API の `rule` が `required`・`maxLength` など） | FE の実装（共通の対応表を 1 か所に置く。概要設計には書かない） |
| 概要設計の相関ルール（`rule` が `x-correlation-rules` の `id`）・業務ルールの違反（`code`） | 概要設計のメッセージ一覧（OpenAPI の `message` が指す `MSG-xx`） |
| 詳細設計だけで決めた相関ルール・業務ルールの違反 | OpenAPI の `x-correlation-rules`・`x-error-codes` の `message` |
| 5xx・通信失敗 | FE の実装（共通の文言） |

Orval は拡張フィールドからコードを生成しないため、コードから文言への対応は FE で手で書く。API が `params` で返した値は文言の `{キー}` に差し込み、表示の形（日時・桁区切りなど）は純関数で整える（[`docs/architecture.md`](../docs/architecture.md) §4）。

> 移行中: public は今、DS の Tailwind テーマと `Button`（`packages/ui`）を使っている。HIG に沿った見た目への置き換え方は未決（§18）。

### Apple HIG（public）

HIG を読むタイミング:

| タイミング | 確認 |
|---|---|
| 新しい画面構成・操作（ナビゲーション、タブ、シート、入力部品など）を初めて作る | 必須 |
| 既存の部品の配置・文言のみ | 原則不要 |
| 挙動・a11y・iPhone での見え方で迷った | 必須 |
| レビューで指摘された | 必須 |

新しい画面構成・部品を作る PR では、参照した HIG のページの URL を記載する。

### デジタル庁デザインシステム（admin）

UIの基本部品はデジタル庁 Design Systemを利用する。

必要な部品を `packages/ui` に取り込む。画面の配色・文字の大きさ・ボタン・確認ダイアログは [管理画面のデザインルール](../docs/rules/ui-design.md) に従う。

usageを確認するタイミング:

| タイミング | 確認 |
|---|---|
| 初めて `packages/ui` に部品を追加する | 必須 |
| 既存部品の配置・文言のみ | 原則不要 |
| 挙動・a11yで迷った | 必須 |
| レビューで指摘された | 必須 |

`packages/ui` に新規部品を追加するPRでは、該当usage URLを記載する。

---

## 5. Componentとロジック

Container / Presenter の専用フォルダや命名は採用しない。

責務で分離する。

| 役割 | 主な置き場 | 責務 |
|---|---|---|
| ロジック | feature `hooks/` | API取得・変換・イベント・画面状態 |
| 表示 | `components/` | propsを受け取り描画 |
| 合成 | `app/` | Router / Provider / featureの組み合わせ |

Component内に以下を溜めない。

- API取得
- 複雑な業務判定
- 大きなデータ変換
- 複雑な状態遷移

必要に応じてfeature hookへ分離する。

---

## 6. API Model と UI

OpenAPI生成型はAPI境界の契約であり、UIの型とは限らない。

原則:

```text
API Model ≠ UI Model ≠ Component Props
```

Presentational ComponentはOpenAPI生成型へ直接依存させない。

API構造とUIで必要な構造が異なる場合のみ、feature hookやmapperでUI向けの形へ変換する。

ただし、単純な受け渡しのためだけにUI ModelやMapperを機械的に作らない。

---

## 7. View State

主要な画面状態は明示する。

例:

```ts
type BooksListViewState =
  | { type: "loading" }
  | { type: "empty" }
  | { type: "error"; message: string }
  | { type: "ready"; items: BookListItem[]; hasMore: boolean }
```

複数booleanの組み合わせで1つの画面状態を表し始めた場合は、`discriminated union` 等によるView Stateを検討する。

---

## 8. 状態管理

状態は最も狭いスコープに置く。

優先順位:

```text
Component local state
→ feature hook
→ Zustand
```

### TanStack Query

サーバ由来データの正とする。

API取得・キャッシュ・再取得・Mutationに利用する。

サーバデータをZustandへ二重保存しない。

### Zustand

複数Component・複数画面をまたぐUI状態に利用する。

単一Componentで完結するmodal open等を機械的にZustandへ置かない。

### Context

Theme / Auth / QueryClientなど横断的な依存注入に限定する。

---

## 9. OpenAPI / Orval

`openapi/dist/<channel>.yaml`（`openapi/` の YAML をバンドルしたもの。BE と同じファイル。public アプリは `public.yaml`、admin アプリは `admin.yaml`）からOrvalで生成する。設定は各アプリの `orval.config.ts`、コマンドは `pnpm gen:api`（全アプリ）。

生成物（各アプリの `src/api/generated/`）:

- API型
- request関数
- TanStack Query hook（名前は operationId から。`useSearchBooks`・`useRegisterBook` など）
- MSW handler（`get<Op>MockHandler`）
- zod スキーマ（`src/api/generated/zod/`。フォームの入力チェック用。制約は OpenAPI の `required`・長さ・範囲・Orval が対応する形式から生成されるので、FE で手書きしない。未対応の形式は OpenAPI に `pattern` などで制約を書く）

生成物は手編集しない。`.gitattributes` の `linguist-generated` で PR の差分では折りたたむ。ESLint / Prettier の対象外。

API の形（必須項目・取りうる値）は OpenAPI の `required` / `enum` で表す。生成型に足りない制約を FE のコードで補わず、OpenAPI を直す（`enum` は BE が値を保証できる項目だけ。[`docs/architecture.md`](../docs/architecture.md) §4）。

フォームの入力チェック（`required`・長さ・範囲・`enum` など）は、Orval が同じ YAML から生成する zod スキーマを使い、FE で値を手書きしない（FE・BE・OpenAPI で制約の値がずれないようにするため）。

基本:

```text
Component
→ Feature Hook
→ Orval Generated Hook
```

ただし、中間hookに意味がない場合は不要なラッパーを作らない。

全リクエストは手書きの `src/api/mutator.ts`（`ErrorResponse` → `ApiError` への正規化）を通る。ローカルでは Vite が `/api` と `/health` を API に中継する（public は :8080、admin は :8081）。

---

### 文字数の共通検証

文字数は Unicode のコードポイントで数える（BE の VO と同じ数え方）。上限値は OpenAPI を正とし、画面ごとに上限を重複定義しない。生成 zod の長さチェックは UTF-16 単位なので、補助平面の文字（絵文字など）を 2 文字と数えてしまう。コードポイント単位の共通処理（例: `Array.from(value).length`）を 1 か所に置き、zod の長さ検証・入力制限・文字数表示はそれを使う。HTML の `maxlength` に依存せず、IME変換確定時・貼り付け・送信時を含めて検証する。上限ちょうどは通り上限+1 は拒否することを、ASCII・日本語・補助平面の絵文字で BE と照合する。結合文字の正規化は文字数検証に暗黙に追加しない。

APIの400では `errors[]` の全件を順序に依存せず表示する。空の `field` は画面全体の違反として `rule` の表示先に従う。該当する入力欄がないためフォーカスしない。共通契約は `docs/architecture.md` §4を参照。

## 10. Scenario / Fixture / MSW

Orval生成MSWはAPI契約のMockとして利用する。生成 API の差し替えは MSW で行い、`vi.mock()` しない（成功は生成ハンドラ、4xx やリクエスト内容の検証は同じメソッド・パスの手書き MSW ハンドラ）。

UIで意味のある状態はScenario / Fixtureとして別途用意する。

例:

- normal
- empty
- manyItems
- longText
- apiError
- slowNetwork

Scenario名はデータ形状ではなく、その状態の意味を表す名前にする。

---

## 11. アプリ内ディレクトリ

各アプリの `src/` は以下を基本とする。

```text
src/
  app/
  api/
  assets/
  components/
  config/
  features/
  hooks/
  lib/
  stores/
  testing/
  types/
  utils/
```

不要なディレクトリは作らない。

feature:

```text
features/<feature>/
  api/
  components/
  hooks/
  stores/
  types/
  utils/
```

必要なものだけ作る。

---

## 12. 依存方向

```text
shared
  ↑
features
  ↑
app
```

feature同士は直接importしない。

必要な場合は以下を検討する。

- `app` で合成
- sharedへ昇格
- 許可された共有物（DS のテーマ・取り込み部品、OpenAPI 生成型、日付の純関数）に限り `packages/` へ昇格

barrel (`index.ts`) は必須にしない。

---

## 13. Storybook

Storybookは必須ではない。

以下の場合に利用を検討する。

- `packages/ui` の独自Component
- Variantが多いComponent
- アプリ上で再現しにくい状態
- Visual Regressionが必要な場合

全ComponentへのStory作成は要求しない。導入したアプリでは CI で `build-storybook` が通ることだけを確かめる。

---

## 14. テスト

目的は件数やCoverageではなく、安心して変更できること。

使用:

- Vitest
- React Testing Library
- MSW

重点的にテストするもの:

- 計算
- Validation
- 業務判定
- Mapper
- View State生成
- Form
- 重要画面

Componentテストはユーザー操作と表示結果を見る。

内部stateやprivate implementationはテストしない。

自作の子Componentを過剰にmockしない。

Snapshotは主力にしない。

Coverage thresholdは設けない。

---

## 15. E2E / Visual Regression

E2E は止まると困る重要導線に限る。現在は Playwright で、本のライフサイクル（登録・編集・削除）と分野タグのライフサイクル（追加・本への付与・分野別の集計への反映・改名・削除）の 2 導線だけを、実バックエンド・実 DB・実フロントエンドサーバーに対して**ローカルでのみ**実行する（CI には載せない。実 DB と 2 つの API サーバーの起動が要り重いため）。分岐網羅・境界値は単体・画面のテストで担保する。

Visual Regressionは必要な画面・Componentが出てから導入する。

---

## 16. CI

PRで原則実行する。

- `typecheck`
- `lint`
- `format:check`
- `test`
- `build`
- `gen:api:check`
- `audit`（high 以上で失敗）

Storybook導入済みの場合のみ `build-storybook` を追加する。

当面必須にしない:

- E2E
- Visual Regression
- Coverage Threshold

---

## 17. 層ごとの規則一覧（決まっていること）

層（ディレクトリ）ごとに散らばっている規則の索引。正は各節（設計は本書、テストの書き方は [テスト規約](../docs/rules/testing.md)）。「実例」は現在のコード。

| 層 | 置き場・ファイル名 | 命名・書き方 | テスト |
|---|---|---|---|
| **app** | `src/app/`。`App.tsx`・`provider.tsx`・`router.tsx`・`routes/XxxRoute.tsx`（実例: `BooksListRoute.tsx`）、admin はログインの振り分け `AdminGuard.tsx` | `App` / `AppProvider` / `AppRouter`。Router / Provider / feature の合成だけを行う（§5） | 合成だけなら個別のテストは書かない（画面のテストで通る）。Provider・Router の挙動（認証による遷移の振り分け等）を足したら、同じディレクトリに `*.test.tsx` を置く |
| **api** | `src/api/`。`generated/`（Orval の生成物。手編集しない）と `mutator.ts` | 全リクエストは `apiMutator` を通り、失敗は `ApiError` に正規化する。`status` は HTTP エラーならそのステータス、通信失敗（オフライン等）なら `0`、成功ステータスで JSON が読めないときは受信したステータス。hook 名・関数名は operationId から生成（`useSearchBooks`）（§9） | `mutator.test.ts`: MSW で成功・204・HTTP エラー・通信失敗・不正な JSON を検証。`zod-constraints.test.ts`: YAML で初めて使う制約の種類ごとに、生成 zod に通る値と弾かれる値の 1 組。生成物そのものはテストしない |
| **features（画面）** | `src/features/<feature>/components/`。画面の表示は `XxxView.tsx`・フォームは `XxxForm.tsx`（実例: `books-list/components/BooksListView.tsx`、`BookForm.tsx`） | Component は PascalCase のファイル・同名の named export（`export function BooksListView`）。API 取得・大きな変換・複雑な状態遷移は hook に出す（§5） | 同名の `*.test.tsx`。MSW の生成ハンドラを `server.use(...)` で上書きし、主要状態（読み込み中 / 0 件 / 成功 / API エラー・通信失敗）ごとに見える内容と押せるものを検証。クエリは `getByRole` 等、操作は `userEvent` |
| **features（hook）** | `src/features/<feature>/hooks/useXxx.ts`（実例: `useBooksList.ts`、`useBookEditor.ts`） | 名前は hook が提供するもので付ける（§17「名前の付け方」）。API の構造と画面が必要な構造が違うときだけ、生成 hook を包んで UI Model に変換する。単純な受け渡しだけなら中間 hook を作らず、生成 hook を直接使う（§6・§9） | ロジックがあれば Component を介さず直接テストする（テスト規約） |
| **features（utils・型）** | `src/features/<feature>/utils/`（実例: `paging.ts`、`barRatio.ts`）、`types.ts` | 関数名と同じ camelCase のファイル名（1 語なら小文字）。型は PascalCase、値の対応表は UPPER_SNAKE（`PAGE_SIZE`） | 計算・判定は直接呼んで分岐・境界値を検証。型だけならテスト不要 |
| **components（共有）** | `src/components/{layouts,pages,states,ui,form}/`。`layouts/` は枠（`Layout`・`Header`・`Footer`）、`pages/` は feature に属さない画面（`NotFoundPage`）、`states/` は読み込み中・0 件・エラーの表示、`ui/` はアプリ内の小さな部品（`RatingStars`・`ConfirmDialog`）、`form/` はフォームの部品（`FormField`・`ErrorSummary`） | 業務の意味を持つ複合 UI は置かない（§2）。DS の部品は `packages/ui`。public の部品は HIG に沿う（§4） | ロジックがあれば同名の `*.test.tsx`。Story は必要なときだけ `*.stories.tsx`（§13） |
| **utils（共有）** | `src/utils/`（実例: エラーのコードから文言を引く `ERROR_MESSAGES` / `getErrorMessage`、コードポイントで数える文字数） | utils は動詞で始まる関数 | 計算・変換は直接呼んで分岐・境界値をテスト |
| **lib / testing** | `src/lib/`（ライブラリの設定。実例: `query-client.ts` の `createQueryClient`）、`src/testing/`（`server.ts`・`handlers.ts`・`render.tsx`・`setup.ts`。E2E は admin の `src/testing/e2e/`） | lib は `createXxx` の生成関数。`testing/` は未定義のリクエストをエラーにする（`onUnhandledRequest: 'error'`）。具体値を見るテストは生成ハンドラを上書きする | — |
| **packages/ui** | `frontend/packages/ui/src/`。部品ごとに PascalCase のファイル、`index.ts` から export | デジタル庁 DS の部品だけ（admin 用。public は移行中のため現在も使用）。新規追加時は usage を読み、PR に URL を書く（§4） | ロジックがあれば同名の `*.test.tsx` |
| **全層共通** | 別ディレクトリの import は `@/` から、同じディレクトリと子ディレクトリは `./`（親ディレクトリへの `../` は ESLint で禁止）。barrel（`index.ts`）は必須にしない（§12） | named export（default export は Storybook の `meta` だけ）。feature 同士は直接 import しない（§12）。名前の付け方は下の「名前の付け方」。コメントは [コメントの書き方](../docs/rules/code-comments.md) | テストは対象と同名で同じディレクトリ。テスト名は日本語で期待する振る舞い。入力違いは `it.each`。Snapshot・Coverage の閾値は使わない（§14） |

### 名前の付け方

命名の規則はこの節が正。名前は目的駆動名前設計に従い、存在や技術的な手段ではなく、何のためのものかから付ける（[`backend/architecture.md`](../backend/architecture.md) §2 の「設計原則」と同じ考え方）。

- Component は何を表示し何を操作できるか、hook は何を提供するか、関数は何をするか、型は何を表すかで名付ける
- 業務の用語（ユビキタス言語: 本・分野タグ・一言まとめ・感想・評価・書影）を揃え、違う業務の概念には違う名前を使う
- `Manager` / `Helper` / `Common` / `Data` / `Processor` / `Util` のような広すぎる名前は、具体的な責務で言い換えられないか見直す（機械的に禁止はしない）
- 短さより、語を足して意味が増えるかで判断する。名前を揃えるためだけに目的の違うものをまとめたり、区別の要らないものを分けたりしない

**表記**

| 対象 | 規則 | 例 |
|---|---|---|
| Component | PascalCase。役割と表示対象を表す（`Card` / `Table` / `Dialog` / `Form` などは実際の役割と合わせる）。`XxxComponent` / `XxxUI` / `XxxManager` にしない | `BookForm`、`ConfirmDialog` |
| 画面 | `XxxPage`（feature に属さない画面）・`XxxRoute`（route の合成）。画面の単位だけに使う | `NotFoundPage`、`BooksListRoute` |
| Component の宣言 | `function` 宣言にし、ファイル名・export 名と揃える。コールバックや短い内部の関数はアロー関数でよい | `export function BookForm(...)` |
| Props の型 | `XxxProps` | `BookFormProps` |
| 型 | PascalCase。`I` の接頭辞を付けない。オブジェクトの型は `interface`、union・交差型・`Record` など `type` でしか書けないものは `type`（ESLint の `consistent-type-definitions` で検査） | `BookFormValues` |
| 関数・変数・props | camelCase。関数は動詞で始め、目的を表す（計算は `calculateXxx`、表示の形に整えるのは `formatXxx`、選ぶのは `resolveXxx`） | `formatPublishedOn` |
| 真偽値 | 意味で `is`（状態）/ `has`（持っている・ある）/ `can`（できる）を使い分ける | `isSubmitting`、`hasMore`、`canDelete` |
| イベントの props | `onXxx` | `onSubmit`、`onClose` |
| Component の中のイベント処理 | `handleXxx` | `handleSubmit` |
| state の setter | `setXxx` | `setValues` |
| 値の対応表の定数 | UPPER_SNAKE | `PAGE_SIZE`、`ERROR_MESSAGES` |
| ほかのモジュールの定数 | 意味に合わせて付ける（すべてを UPPER_SNAKE にはしない） | `defaultPageSize` |

**hook**

- hook が外に提供する状態・情報・操作で名付ける。画面名・API の関数名・実装の技術から機械的に付けない（`useBooksList` / `useBookEditor` / `useLoginForm`）
- `Query` / `Mutation` の接尾辞は、区別が要るときだけ付ける
- 生成 hook は operationId の名前のまま使い、名前を変えるためだけの中間 hook を作らない（§9）
- operationId は BE の UseCase の名前に揃える（`SearchBooksUsecase` → `searchBooks` → `useSearchBooks`。[`docs/architecture.md`](../docs/architecture.md) §4）。FE の独自 hook の名前は operationId に揃えようとしない

**API Model → UI Model の変換（mapper）**

- 変換先が 1 つで明らかなら `toXxx`、変換元・変換先が曖昧なら `mapXxxToYyy`。`convertData` / `adapt` のような抽象的な名前にしない
- UI Model は画面での目的・意味で名付け、`Model` / `View` の接尾辞は必須にしない。API Model と同じ形で変換が要らなければ mapper を作らない（§6）

**Zustand の store**

- hook の名前は `useXxxStore`、ファイルも同じ名前。状態の型は `XxxState`、操作は具体的な動詞
- 置き場は最も狭いスコープから選ぶ（§8）。store の名前を決めるためだけに store を作らない

**feature**

- `src/features/<feature>/` は業務の領域か、利用者に提供する機能で名付ける（`books-list` / `book-detail` / `tag-counts`）
- `home` のような画面名は、その画面の表示と合成が固有の責務であるときだけ使う
- `common` / `misc` のように、何が属するかの基準が曖昧な名前を作らない

**ファイル**

| 種類 | 規則 | 例 |
|---|---|---|
| Component・画面・Story・テスト | export する Component の名前と同じ PascalCase | `BookForm.tsx`・`BookForm.test.tsx`・`BookForm.stories.tsx` |
| hook・store | hook の名前と同じ camelCase | `useBooksList.ts` |
| utils の関数 | 関数名と同じ camelCase（1 語なら小文字） | `barRatio.ts`・`paging.ts` |
| ライブラリの設定など、それ以外の複数語の `.ts` | kebab-case | `query-client.ts` |

JSX を含むファイルは `.tsx`、含まないファイルは `.ts`。

**テスト用のデータ・その他**

- Scenario は状態の意味で名付ける（§10）。fixture を作る関数は `createXxxFixture`
- エラーのコードから文言を引く対応表は `ERROR_MESSAGES`、引く関数は `getErrorMessage`。API の通信エラーは `ApiError` に正規化し、Component ごとにエラーの型を増やさない

---

## 18. 未決の規則（必要になったときに決める）

まだ実例が無い、または現在のコードで書き方が揃っていないもの。該当する実装が初めて出る PR で決め、決めたら §17 と各節に反映する。推測で先に決めない。

| 対象 | 決めること |
|---|---|
| feature 内の `api/` | 生成 hook を包む関数を feature の `api/` に置くか（§11 に枠だけあり、実例なし） |
| Scenario / Fixture | §10 の Scenario・fixture の置き場と、どこまで再利用するか（現在はテストごとにその場で書いている） |
| ルーティング | path の定義の置き場（現在は `/books/${id}` を Component に直接書いている）と、画面と route の対応の書き方 |
| hook のテスト | `renderHook` で hook を直接テストする書き方と、生成 hook の差し替え方 |
| 依存の機械的な検査 | 次は ESLint（`eslint.config.mjs`）で検査済み: feature 間の import 禁止・共有コードから features への import 禁止・別ディレクトリの相対 import（`../`）禁止・default export 禁止（Story を除く）。feature・共有のディレクトリを足したら設定にも足す。ファイル・識別子の命名と `function` 宣言を検査するかは未決 |
| HIG に沿った部品（public） | 部品の置き場（アプリの `components/ui/` か共有パッケージか）／HIG に沿った Web 向け UI ライブラリを使うか自作するか |
| public の DS からの移行 | 今使っている DS の Tailwind テーマ・`Button` を HIG に沿った見た目に置き換える手順と時期 |
