# テストの書き方・置き場所

本プロジェクトのテスト規約。フロントの「何をどの粒度でテストするか」（単体 / 画面 / E2E の使い分け・鉄則）は
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library) を正とし、
本書は backend / frontend 共通の「どう書くか」と、スタック別の「どこに置くか・何で差し替えるか」だけを定める。

## 全スタック共通: どう書くか

以下は backend（Go）/ frontend（React）のいずれでも守る。
言語やフレームワークが変わっても、テストが担保すべき観点は変わらない。

- **分岐網羅**: 条件分岐は真偽の両方を通す。早期 return・ガード節は、そこで止まる入力と
  通り抜ける入力の両方をテストする。片方だけでは分岐の半分が未検証のまま残る
- **境界値**: 長さ・件数などに上限がある項目は「上限ちょうど」と「上限+1（超過）」の
  2 ケースを書く。どちらか片方だけでは、`<` と `<=` の取り違えのような誤りを検出できない
- **正規化の検証**: トリム・空文字→null 変換などの正規化ロジックがあるなら、
  正規化前の入力と正規化後の値の両方をアサーションする
- **失敗時の非変化**: 更新系の操作がエラーで失敗したとき、対象の状態が変わっていない
  （呼び出し前の値のまま）ことを検証する
- **エラーの伝播**: 下位層（Repository / Query・外部API等）が返すエラーが、上位層で握り潰されず
  そのまま（または意図した形で）伝播することを検証する。バリデーションエラー時は
  下位層が呼ばれないことも検証する
- **テスト名は「期待する振る舞い」を書く**: `CLAUDE.md` の「Test code writes **What**」に
  対応する。実装の手順ではなく仕様を書く（例: `titleが101文字ならエラー`）
- **同じ検証を入力だけ変えて繰り返す場合はテーブル化する**（Go のテーブルドリブンテスト、
  Vitest の `it.each` 等、各言語の標準機能を使う）

## カバレッジについて数値目標は設けない

カバレッジ率の閾値は導入しない。数値目標は「分岐を通しているがアサーションが無い」
「境界値を検証していない」ような意味の薄いテストでも達成できてしまい、
かえって上記の観点を形骸化させる。

## フロントエンド: 描画と動作の観点

主要状態（ローディング / 成功 / 0件 / 入力エラー / API エラー / 操作成功）の確認は
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library) に従う。
それに加えて、モーダル・ドロップダウンの開閉や画面遷移など「ロジックとしては単純だが、ユーザー操作に対する
見え方・遷移が仕様通りか」を独立した観点として検証する。ロジックが正しくても UI に反映されていなければ
ユーザー体験としては壊れているため、バリデーション成否等のロジック観点とは別にアサーションする。

- **開閉状態**: `open` が false なら描画されない／true なら中身が表示される
- **操作と結果の対応**: どの操作でどのコールバックが呼ばれるか。「確定は `onConfirm` だけ、キャンセルは
  `onCancel` だけ」のように、意図しない方が呼ばれていないことまで検証する
- **画面遷移**: 認証状態などの条件によって画面がどこへ遷移するか

**ビジュアル観点（色・サイズ・配置・レスポンシブ崩れ等）は自動テストの対象外**。
Storybook は表示バリエーションのカタログであり、テストの代わりにしない（視覚回帰は当面導入しない。
[`frontend/architecture.md` §3.1・§3.2](../../frontend/architecture.md#31-storybookpresentational-の状態カタログ)）。
見た目は PR の「確認事項」にあるローカル UI 確認（人の目視）で担保する。

## スタック別: どこに置くか・何で差し替えるか

### backend（Go）

層・ディレクトリ構成は [`backend/architecture.md`](../../backend/architecture.md) を正とする。

- テストは対象と**同じディレクトリ**に `*_test.go` として置く
  （例: `internal/domain/book/price.go` と `price_test.go` が同居）
- **原則、外部テストパッケージにする**: package 名は `<対象>_test`（例: `package book_test`）。
  公開 API を通した振る舞いを検証する
- **例外**: 非公開シンボルの検証が必要な場合のみ in-package（`package <対象>`）を許可。
  なぜ外から検証できないのかをファイル冒頭のコメントに残す
- 入力違いの検証はテーブルドリブン（`tests := []struct{...}` + `t.Run`）で書き、`t.Parallel()` を付ける
- テストダブルは**手書き Fake**（domain の `Repository` / `Query` IF を満たす構造体を手で書く）を使う。
  UseCase のテストは Fake を注入し、infrastructure（PostgreSQL）に繋がずに走ることを目標にする
- Handler のテストは `httptest` + `common.NewEcho()` を使い `e.ServeHTTP` の実リクエストで検証する
  （Handler を直接呼ぶだけだとエラーが応答へ変換されない。変換はルーターに登録された
  `common.HTTPErrorHandler` が担うため）。`BindValidate` による 400（フィールドエラー含む）と、
  domain error → HTTP ステータスの変換を確認する
- **Repository／Query／ExternalGateway の実装（Infrastructure）は契約テストで検証する**: 実際の
  PostgreSQL（`docker compose`）や外部サービスの mock に対して実行し、SQL・DB マッピング・
  API リクエスト/レスポンス・シリアライズ・トランザクションなど、Domain の IF を満たすこと・
  期待する読み書き結果を確認する。UseCase 側の Fake テストとは目的が異なり（Fake はシナリオ検証、
  契約テストは実装が IF 通りに動くことの担保）、どちらかで代替しない
- **参照系（Query）のテストは Read Model へのマッピング固有の観点も確認する**: 必要な情報が
  取得できること・Read Model の各フィールドへ正しくマッピングされることに加えて、**更新系の
  Domain Model（Entity・VO）に依存していないこと**（例: Read Model の `Title` が `string` であり
  書き込み側の `Title` VO 型を返していないこと）を確認する
- **Presentation → Infrastructure の結合テスト**: 各 Handler につき、`httptest` + `common.NewEcho()` +
  実際の DI 配線（`registerRoutes` 相当）+ 実際の PostgreSQL を通した実 HTTP リクエストで、その Handler が
  **返しうる HTTP ステータスごとに最低 1 件**のテストを書く（200 のハッピーパス＋実際に返しうる 4xx/5xx を
  それぞれ 1 件ずつ）。狙いは内部の分岐網羅ではなく、Handler→UseCase→Repository/Query→DB の配線が
  実物同士で噛み合っていることの確認であり、既存の Fake ベースの Handler／UseCase テストと契約テストを
  代替しない（分岐網羅・境界値はそれらに任せる）。外部 API（openBD・楽天など）に依存する Handler は、
  その依存だけ Fake の Gateway に差し替え、DB は実物のままにする（外部ゲートウェイ自体は
  `infrastructure/gateway/<name>` の単体テストで別途検証済みのため、ここで実ネットワークに頼る必要はない）。
  ファイルは対象と同じディレクトリに `<handler>_integration_test.go` として置き、DB に繋がらなければ
  契約テストと同じ理由（`docs/rules/testing.md` の Repository／Query の項を参照）で skip する
- 実行: `cd backend && make test`（= `go test ./...`）

### frontend（React + Vite）

ランナー・ライブラリ・置き場（コロケーション / `src/testing/`）は
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library) を正とする。
ここでは書き方の細則だけを定める。

- テストは実装と**同名**の `*.test.ts(x)` として同じディレクトリに置く
  （例: `features/home/components/HomePage.tsx` と `HomePage.test.tsx`）
- **ロジックはコンポーネントを介さず単体でテストする**: バリデーション・金額や日付の計算・hooks・store は
  レンダリングせずに直接呼んで分岐網羅・境界値を検証する。コンポーネント側のテストは
  「入力するとエラー文言が表示され送信されない」という**配線**の確認に絞り、ロジックの全分岐を
  コンポーネント経由で網羅し直さない
- テストダブルは HTTP 境界で差し替える（MSW、または feature の `api/` モジュールを `vi.mock()`）。
  自作の子コンポーネントはモックしない
- 操作は `userEvent` を使う（`fireEvent` は使わない）。非同期のアサーションは `findBy` / `waitFor` を
  優先し、`getBy` + 手動の `act` は避ける
- **クエリはアクセシブルな属性を優先する**: `screen.getByRole` / `getByLabelText` / `getByText`
  で要素を問い合わせる（実装の詳細ではなく、ユーザーが実際に見る画面の見え方を検証するため）。
  `getByTestId` は上記のクエリで届かない場合の最終手段
- 実行: `cd frontend && pnpm test`
