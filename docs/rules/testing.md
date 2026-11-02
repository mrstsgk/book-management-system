# テストの書き方・置き場所

本プロジェクトのテスト規約。フロントの「何をどの粒度でテストするか」（単体 / 画面 / E2E の使い分け・鉄則）は
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library) を正とし、
本書は backend / frontend 共通の「どう書くか」と、スタック別の「どこに置くか・何で差し替えるか」だけを定める。

## 全スタック共通: どう書くか

以下は backend（Go）/ frontend（React）のいずれでも守る。
言語やフレームワークが変わっても、テストが担保すべき観点は変わらない。

- **正しい振る舞いで成功し、検出対象の誤った振る舞いで失敗することを確認する**:
  新規開発・機能変更・バグ修正を問わず適用する。テストが通ることだけでなく、
  どの誤実装を検出するテストなのかを確認する
- **テストの入力値・期待値を検証対象の実装から独立させる**:
  検証対象の実装定数やロジックから導出せず、仕様に基づいて独立して定義する。
  実装の誤変更にテストが追従してしまわないようにする
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

### 実装の誤りを検出できるかの確認

例えば仕様が「感想は 5000 文字まで」なら、テストには仕様に基づく値として「5000 文字は成功、5001 文字は超過で失敗」を書く。
実装の `MaxCommentLength` と `MaxCommentLength + 1` から入力を作るだけでは、実装の定数を誤って 5001 に変更しても
テストが追従して成功するため、仕様の上限が 5000 文字であることを保証できない。
テスト側で仕様の値を表す定数を定義してもよいが、検証対象の実装定数を参照しない。

作成時・レビュー時には、期待する結果のアサーションがあり、検証対象の処理を実際に通っていることに加え、
条件の反転・必要な処理の削除・上限値の変更などで失敗するかを確認する。
判断が難しい場合は一時的に実装へその誤りを入れてテストを実行し、意図したアサーションで失敗することを確かめてから戻す。
構文エラーや環境不備による失敗を、誤実装を検出できた証拠にしない。

## 正常系だけで終わらせない: 異常系の観点

正常系のテストだけでは、壊れやすい箇所（不正な入力・競合・失敗）が検証されないまま残る。
テストを書く前に、対象の変更に次の観点が当てはまるかを確認し、当てはまるものは必ずテストにする
（要件定義（[`docs/specifications.md`](../specifications.md)）・機能ごとの設計・API が返す 4xx/5xx から洗い出す。実装の今の動きから書かない）。
該当しない観点は書かなくてよい。

| 観点 | 確認すること |
|---|---|
| 入力バリデーション（準正常系） | 必須の欠落・型違い・形式違い（ISBN の桁・ハイフン）・長さ／範囲の外（感想の文字数、評価の 1〜5）。利用者が起こしうる入力ミスで、システムの異常ではない。違反ごとに期待するエラー（ステータスと `errors[{field, rule}]`）を検証する。検証は層ごとに持つ: HTTP 入力は Request DTO の `validate` タグ（`BindValidate` の変換は `presentation/http/common` のテストで網羅し、各 DTO に書いた制約の値は Handler のテストで検証）、不変条件は Domain の VO のコンストラクタ（Domain のテスト） |
| 業務ルール違反（準正常系） | 入力は正しいが業務上できない操作（同じ ISBN の二重登録・openBD に無い ISBN・タグ名の重複・対象が存在しない）。違反ごとに期待するエラー（`ErrInvalid` / `ErrNotFound` / `ErrConflict` とそれに対応する 4xx）になり、状態が変わらない |
| 同値分割・デシジョンテーブル | 結果が複数の条件の組み合わせで決まるとき（例: 書名の上書きがあるか × 外部カタログを取り直せたか）、意味のある組み合わせごとにテストがある。全部が真のケースだけにしない |
| 状態遷移 | ステータスや版を持つものは、不正な遷移が拒否されること（例: 期限切れセッションでの操作、ログアウト済みセッションの再利用） |
| 同時実行 | 楽観ロックの version 不一致で後から保存した方が失敗すること、ログイン失敗の連続で一時的に拒否されること（実際の PostgreSQL で検証する） |
| 冪等性 | 再送・二重送信をしても結果が同じで、副作用が重複しない（例: ログアウトの二重送信） |
| 認可 | 未ログインのリクエストと、Origin が一致しない書き込みが拒否される |
| 日時 | セッションの期限ちょうど、日付またぎ、JST と UTC |
| 外部障害 | openBD のタイムアウト・エラー応答・部分的な欠落（書影なし）で、状態が中途半端に更新されない（更新時に取り直せなければ今の書誌のままにする） |
| 空・重複 | 0 件、空文字、null、重複したタグ ID、一覧の最後のページ |

層ごとの最低ラインは次のとおり。

- **Domain**: 不正な入力でオブジェクトを作れないこと、不正な状態遷移の拒否
- **UseCase**: 下位層のエラーの伝播、失敗時に状態が変わらないこと、入力エラー時に下位層が呼ばれないこと
- **Presentation**: 返しうる 4xx/5xx ごとに 1 件と、Request DTO の `validate` 制約ごとに通る値と弾かれる値の 1 組
- **フロント**: API エラー・入力エラーの状態（成功の画面だけにしない）

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

## フロントエンド: 何をテストするか（一覧）

何を・どこで・どう検証するかの早見表。詳細は上の各節と
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library)。

| 対象 | テストする | 書き方 |
|---|---|---|
| 計算・変換・並び替え・判定などのロジック | する | コンポーネントを介さず直接呼び、分岐網羅・境界値を検証する。コンポーネントの中に埋まっていたら、純関数か hook に出してからテストする |
| 画面の表示 | する | 主要状態（読み込み中・0 件・成功・API エラー）ごとに、見える内容と押せるものをユーザー操作で検証する |
| 開閉・操作と結果の対応・遷移 | する | 上の「描画と動作の観点」 |
| 入力チェックのロジック（`validateBookForm` のような純関数） | する | 単体で分岐網羅・境界値を検証する。制約の値は BE の Handler のテスト（`validate` タグ 1 つにつき通る値と弾かれる値）と揃える。画面では「違反時にエラーが表示され、送信されない」という配線だけを検証する |
| サーバーが返した入力エラー（400） | する | MSW で 400 を返し、`errors` の `field` に対応する入力欄のすぐ近くに表示されること（`errors` を含まない 400 は画面上部）と、API の既定のメッセージがそのまま出ないことを検証する |
| API 通信とエラーの正規化 | する | `mutator` のテストで網羅する。画面では表示だけを検証する |
| API Model → UI Model の変換 | 変換があるときだけ | 単体で検証する |
| 生成物（API 型・hook・MSW） | しない | 生成物を**使う側**（hook・画面）のテストは必須 |
| 見た目（色・サイズ・配置） | しない | PR の目視確認で担保する |
| 内部 state・自作の子コンポーネントの mock・Snapshot | しない | 上のとおり |

同じ検証を層をまたいで繰り返さない。制約の値は BE、変換とロジックは単体、画面は配線と表示、と分担する。

## 生成物とテスト・レビュー

生成物（swag が排出する `backend/api/docs/`、Orval が生成する `frontend/web/src/api/generated/`）は
**テストもレビューもしない**（CI のドリフト検査は、生成物が今のコードとツールで生成した結果と同じかだけを見る。
生成ツールの変換の正しさは、版の固定と、生成物を使う側のテストで間接的に担保する）。一方、生成物を**実装・利用する
コード（BE の Handler、FE の hook・画面）のテストは必須で、そのテストはレビュー対象**とする。

## テストの有無の検査（CI: `Test Presence`）

テストの書き忘れに気づくため、PR で次を検査する（`.github/workflows/test-presence.yml`）。

| 検査 | 内容 | 結果 |
|---|---|---|
| API の operation ごとのテスト | operation の一覧は `backend/api/docs/swagger.json` の `paths` から取る（CI のドリフト検査でコードと一致している）。**BE**: `backend/internal/presentation/http/` 配下の `*_test.go` のいずれかに、そのパスの文字列（`{id}` は任意の 1 セグメント）と HTTP メソッド（`http.MethodGet` または `"GET"`）の両方があること（リクエストを組み立てる補助関数にメソッドが隠れるため、同じ行には限定しない）。**FE**: アプリのコード（生成物・テスト・Story・`src/testing/` 以外）が生成モジュール `api/generated/api` からその operation の hook・関数（`use<Op>`・`use<Op>Infinite`・`get<Op>QueryOptions`・`get<Op>QueryKey`・`<op>`。`<Op>` は Orval の命名で HTTP メソッド + パス、例 `GetApiBooks`）を使っていれば、`*.test.ts(x)` のいずれかが生成 MSW ハンドラ `get<Op>MockHandler(` を呼ぶか、同じメソッド・パスの手書き MSW ハンドラ（`http.put('*/api/tags/:id', …)`。4xx やリクエスト内容の検証に使う）を使っていること（import だけでは不可。まだ使っていない operation は対象外）。テスト・ソースはコメントを除いて判定する。文字列での判定のため、別の operation のテストに同じパスが現れると一致する限界がある | **失敗** |
| 変更したソースのテスト | PR で変更した `backend/{internal,cmd}/**/*.go`・`frontend/*/src/**/*.ts(x)`（`frontend/packages/*/src` を含む。テスト・Story・`*.d.ts`・`src/testing/`・生成物を除く）のうち、同じディレクトリのテストが追加・変更されていないもの。GitHub の注釈は 1 ステップ 10 件までしか表示されないため、全件はジョブのサマリーに出す | 警告のみ |

失敗にするのは、「API があるのにテストが無い」は客観的に判定でき例外が無いため。変更したソースのテストは、要らない正当な変更があり機械的に判定できないため警告にとどめる（失敗にすると形だけのテスト変更を誘う）。
正は CI（人が書いた PR も検査するため）。Stop hook `require-tests-on-change.sh`（変更したソースと同名のテストの変更が無ければ止める）は AI の作業時の先行チェックとして残すが、判定は CI と同じではなく、hook が通っても CI で失敗しうる。

警告はリファクタリングなどテストの変更が要らない変更でも出る。振る舞いが変わる変更なら、テストを足してから
レビューに出す。手元では `.github/scripts/check-api-tests.sh`・`.github/scripts/warn-missing-tests.sh origin/develop` で同じ検査ができる。

## スタック別: どこに置くか・何で差し替えるか

### backend（Go）

層・ディレクトリ構成は [`backend/architecture.md`](../../backend/architecture.md) を正とする。

#### 層ごとのテストと担保すること

詳細は下の各項目。同じことを 2 つの層で検証しない（下の層で担保したことは上の層で繰り返さない）。

| 層 | テスト（ファイル） | 何をテストするか | そのテストが担保すること |
|---|---|---|---|
| Domain | `internal/domain/<domain>/*_test.go` | VO・Entity のコンストラクタと業務ルール。DB・HTTP に依存しない | 業務上の制約の境界値（要件定義から）・不変条件・状態遷移。不正な状態を作れないこと |
| UseCase | `internal/usecase/<domain>/{command,query}/*_test.go` | 手書き Fake を注入したシナリオ | 業務上の認可・分岐・下位層のエラーの伝播・失敗時に状態が変わらないこと・入力エラー時に下位層を呼ばないこと |
| Infrastructure（DB） | `internal/infrastructure/postgres/<domain>/*_test.go` | 実際の PostgreSQL に対する契約テスト。**SQL が複雑なときに書く** | JOIN・集計・ロック・トランザクション・DB 固有の機能が期待どおりに読み書きすること |
| Infrastructure（外部連携） | `internal/infrastructure/gateway/<提供元>/*_test.go` | 外部サービスの mock（`httptest`）に対する契約テスト | リクエスト・レスポンスの組み立てとシリアライズ |
| presentation/http/common | `internal/presentation/http/common/*_test.go` | `BindValidate` の変換とエラーの変換 | 制約の**種類**（`required`・`max`・`min`・`oneof` 等）ごとに違反が 400 + `errors[{field, rule}]` になること。domain error → HTTP ステータスの変換 |
| Presentation（Fake） | `internal/presentation/http/<domain>/handler_test.go` | Fake の UseCase を注入し `e.ServeHTTP` で実リクエスト | Request DTO に書いた `validate` 制約がすべて要件どおりであること（制約 1 つにつき通る値と弾かれる値の 1 組）・UseCase への受け渡し・domain error → ステータス |
| Presentation（結合） | `internal/presentation/http/<domain>/handler_integration_test.go` | 実際の DI 配線 + 実際の PostgreSQL で、API が返しうるステータスごとに実リクエスト | Handler → UseCase → Repository／Query → DB の配線が実物同士で噛み合っていること・DB の値が期待するフィールドに載ること |
| cmd | `cmd/api/main_test.go` | `registerRoutes` で組み立てたサーバーへの実リクエスト | ルートが登録され依存に nil が無いこと・セッション／同一オリジンの middleware が書き込み系に適用されていること |

- テストは対象と**同じディレクトリ**に `*_test.go` として置く
  （例: `internal/domain/book/title.go` と `title_test.go` が同居）
- **原則、外部テストパッケージにする**: package 名は `<対象>_test`（例: `package book_test`）。
  公開 API を通した振る舞いを検証する
- **例外**: 非公開シンボルの検証が必要な場合のみ in-package（`package <対象>`）を許可。
  なぜ外から検証できないのかをファイル冒頭のコメントに残す
- 入力違いの検証はテーブルドリブン（`tests := []struct{...}` + `t.Run`）で書く。`t.Parallel()` は、DB・環境変数・グローバル状態など他のテストと共有するものを使わないテストにだけ付ける（共有するものを使うテストを並列にすると、競合して結果が不安定になる）
- テストダブルは**手書き Fake**（domain の `Repository` / `Query` IF を満たす構造体を手で書く）を使う。
  UseCase のテストは Fake を注入し、infrastructure（PostgreSQL）に繋がずに走ることを目標にする
- **Domain のテストは DB・HTTP に依存しない**: 不変条件・業務ルール・状態遷移・境界値、および不正な状態を
  生成できないこと（コンストラクタがエラーを返すこと）を検証する
- Handler のテスト（`handler_test.go`）は `httptest` + `common.NewEcho()` を使い `e.ServeHTTP` の実リクエストで検証する
  （Handler を直接呼ぶだけだとエラーが応答へ変換されない。変換はルーターに登録された
  `common.HTTPErrorHandler` が担うため）。`BindValidate` による 400（フィールドエラー含む）と、
  domain error → HTTP ステータスの変換を確認する。**Request DTO に書いた `validate` 制約がすべて要件どおりか**は、
  operation ごとに 1 つのテーブルドリブンのテストにまとめ、制約 1 つにつき「通る値」と「弾かれる値（400 と `errors[{field, rule}]`）」の 1 組を
  要件定義から取って送る（制約の値は BE の `validate` タグと FE の入力チェックの両方の元になり、書き間違いはこれ以外のテストでは見つからない）

    | 制約 | 通る値 | 弾かれる値 |
    |---|---|---|
    | `required` | 値あり | 欠落 |
    | `min`・`max`・`gte`・`lte` | 境界ちょうど | 境界の 1 つ外 |
    | `oneof` | 集合の各値 | 集合の外の値 |
    | `dive`・配列の件数 | 件数の境界ちょうど | 境界の 1 つ外 |
    | 形式（ISBN 等の独自検証） | 合う値 | 合わない値 |

- 制約の種類ごとの変換（`validate` のタグ → 400 + `errors[{field, rule}]`）は `presentation/http/common` のテストで網羅する。
  Handler のテストでは種類ごとの網羅を繰り返さない。**初めて使う種類の制約は、その PR で common のテストに足す**
- **Repository／Query の契約テストは、SQL が複雑なときに書く**: JOIN・サブクエリ・集計（`GROUP BY` 等）・
  ロックや複数文のトランザクション（楽観ロックの version 比較など）・DB 固有の機能を使う場合が対象。
  単純な取得・保存だけなら必須にしない（SQL と詰め替えは結合テストで実際の DB を通る。既存の契約テストは残してよい）
- **ExternalGateway の実装（Infrastructure）と、上の条件に当たる Repository／Query は契約テストで検証する**: 実際の
  PostgreSQL（`docker compose`）や外部サービスの mock に対して実行し、SQL・DB マッピング・
  API リクエスト/レスポンス・シリアライズ・トランザクションなど、Domain の IF を満たすこと・
  期待する読み書き結果を確認する。UseCase 側の Fake テストとは目的が異なり（Fake はシナリオ検証、
  契約テストは実装が IF 通りに動くことの担保）、どちらかで代替しない
- **参照系（Query）の契約テストを書く場合は、Read Model へのマッピング固有の観点も確認する**: 必要な情報が
  取得できること・Read Model の各フィールドへ正しくマッピングされることに加えて、**更新系の
  Domain Model（Entity・VO）に依存していないこと**（例: Read Model の `Title` が `string` であり
  書き込み側の `Title` VO 型を返していないこと）を確認する
- **Presentation → Infrastructure の結合テスト**: 各 Handler につき、`httptest` + `common.NewEcho()` +
  実際の DI 配線（`registerRoutes` 相当）+ 実際の PostgreSQL を通した実 HTTP リクエストで、その Handler が
  **返しうる HTTP ステータスごとに最低 1 件**のテストを書く（200 のハッピーパス＋実際に返しうる 4xx/5xx を
  それぞれ 1 件ずつ）。狙いは内部の分岐網羅ではなく、Handler→UseCase→Repository/Query→DB の配線が
  実物同士で噛み合っていることの確認であり、既存の Fake ベースの Handler／UseCase テストと契約テストを
  代替しない（分岐網羅・境界値はそれらに任せる）。外部 API（openBD など）に依存する Handler は、
  その依存だけ Fake の Gateway に差し替え、DB は実物のままにする（外部ゲートウェイ自体は
  `infrastructure/gateway/<name>` の単体テストで別途検証済みのため、ここで実ネットワークに頼る必要はない）。
  ファイルは対象と同じディレクトリに `<handler>_integration_test.go` として置く。500 はキャンセルした context で
  リクエストを送り、実際の DB の呼び出しを失敗させて作る（DB を止めない）
- **`cmd` の配線のテスト**: `cmd/api/main_test.go` で、`registerRoutes` に実際の依存を渡し、各ルートが Handler → UseCase を通って
  応答すること（依存の nil・登録漏れ）と、書き込み系にセッション・同一オリジンの middleware が適用されていることを確認する。
  `cmd` は `package main` なので in-package で書く
- **DB を使うテスト（契約・結合・`cmd` の配線）の準備**: 接続先はローカルの docker compose の既定値。
  テーブルは `make migrate-up` 済みが前提（CI は Backend CI の `check` ジョブで PostgreSQL を起動して migration を当てる）。
  データはテストの終わりに片付ける（他のテストに残さない）。**DB に繋がらなければ、ローカルでも CI でも skip せず失敗にする**
  （skip のまま通ると何も検証しないため。`make test` の前に `make db-up && make migrate-up`）
- 実行: `cd backend && make test`（= `go test -p 1 -count=1 ./...`。DB の状態はテストキャッシュの入力に含まれないため、キャッシュを使わない。DB を使うテストは別のパッケージ同士でも同じテーブルのデータを作って消すため、パッケージを並行に実行しない。CI も `make test` を使う）

### frontend（React + Vite）

ランナー・ライブラリ・置き場（コロケーション / `src/testing/`）は
[`frontend/architecture.md` §8](../../frontend/architecture.md#8-テストvitest--react-testing-library) を正とする。
ここでは書き方の細則だけを定める。

- テストは実装と**同名**の `*.test.ts(x)` として同じディレクトリに置く
  （例: `features/books-list/components/BooksListView.tsx` と `BooksListView.test.tsx`）
- **ロジックはコンポーネントを介さず単体でテストする**: バリデーション・金額や日付の計算・hooks・store は
  レンダリングせずに直接呼んで分岐網羅・境界値を検証する。コンポーネント側のテストは
  「入力するとエラー文言が表示され送信されない」という**配線**の確認に絞り、ロジックの全分岐を
  コンポーネント経由で網羅し直さない
- テストダブルは HTTP 境界で差し替える。**生成 API（`src/api/generated`）は MSW で差し替え、`vi.mock()` しない**。成功の応答は Orval が生成した MSW ハンドラ（`get<Op>MockHandler`）を使い、4xx やリクエスト内容の検証は同じメソッド・パスの手書き MSW ハンドラ（`http.put('*/api/tags/:id', …)`）で行う（API 契約の Mock として扱うため。CI `Test Presence` もこの形を検出する）。`vi.mock()` は生成物以外の feature の `api/` モジュールに限る。
  自作の子コンポーネントはモックしない
- 操作は `userEvent` を使う（`fireEvent` は使わない）。非同期のアサーションは `findBy` / `waitFor` を
  優先し、`getBy` + 手動の `act` は避ける
- **クエリはアクセシブルな属性を優先する**: `screen.getByRole` / `getByLabelText` / `getByText`
  で要素を問い合わせる（実装の詳細ではなく、ユーザーが実際に見る画面の見え方を検証するため）。
  `getByTestId` は上記のクエリで届かない場合の最終手段
- 実行: `cd frontend && pnpm test`
