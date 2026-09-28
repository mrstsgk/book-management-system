# 読んだ本の API 設計（既存）

**日付:** 2026-09-28
**対象:** 読んだ本の登録・更新・削除・取得・一覧と、登録前のカタログ確認（#58〜#60、#63・#64 で削った後の形）

この文書は実装済みの機能を、あとから機能ごとの設計として書き起こしたもの。以後の機能（書名の上書き・一言まとめ など）の設計書は、ここからの差分として書く。

| 別の文書に書いてあること | 文書 |
|---|---|
| API の一覧と認証の要否 | [`backend/architecture.md`](../../../backend/architecture.md) §4 |
| 書影を外部カタログの URL で見せる理由 | [ADR](../../adr/2026-09-28-book-cover-from-external-catalogs.md) |
| 楽天を書影の補完だけに使う理由 | [ADR](../../adr/2026-09-28-rakuten-for-cover-only.md) |
| 作り直した理由 | [ADR](../../adr/2026-09-28-rebuild-as-reading-portfolio.md) |
| テーブルの定義 | [`docs/db/backend-schema.md`](../../db/backend-schema.md) |

## 1. ドメイン（`domain/book`）

### 1.1 集約と値オブジェクト

| 型 | 種類 | 規則 |
|---|---|---|
| `Book` | 集約 | `ID`・`ISBN`・`Bibliography`・`Cover`（無い本は nil）・`Comment`・`Rating`・`Version` を持つ |
| `ISBN` | VO | 13桁または10桁（ハイフン可）。常に13桁で持ち、10桁は 978 付きに変換する。ISBN-10 の形式（978 で始まるものだけ）と Amazon の商品ページ URL を導出できる |
| `Bibliography` | VO | 書名（1〜255 文字、必須）・著者（500 文字まで）・出版社（255 文字まで）・発売日（32 文字まで）。前後の空白を除き、制御文字は受け付けない |
| `Cover` | VO | 書影の URL（https）と提供元（`openbd` / `rakuten`）。画像は保存しない |
| `Comment` | VO | 感想。前後の空白を除いて 1〜5000 文字。改行とタブは使える |
| `Rating` | VO | 評価。1〜5 の整数 |

`Book` の操作:

- `New`: 登録前の本を作る（ID と版は保存時に採番する）
- `ChangeReview(comment, rating, version)`: 感想と評価を差し替える。`version` は更新元が読んだ版（楽観的ロック）
- `RefreshCatalog(bibliography, cover)`: 外部カタログから取り直した書誌と書影に差し替える

### 1.2 ポート（IF）

| IF | 種類 | 役割 |
|---|---|---|
| `Repository` | 書き込み側 | `FindByID`・`Create`（同じ ISBN は `ErrConflict`）・`Update`（版が違えば `ErrConflict`）・`Delete` |
| `Query` | 参照側 | `FindDetailByID`（`BookDetail`）・`FindList(ListRange)`（`BookList`: 項目と総件数） |
| `BookCatalog` | ExternalGateway | `Lookup(ISBN)` で書誌と書影（`CatalogEntry`）を返す。該当なしは `ErrNotFound` |

Read Model（`BookDetail`・`BookListItem`）は文字列と数値だけを持ち、VO 型を返さない。Amazon の URL は ISBN から導出して載せる（保存しない）。一覧の項目は感想の本文を持たない。

`ListRange`（`domain/common`）: 件数 1〜100（既定 20）、開始位置 0 以上。

## 2. 外部カタログ（`infrastructure/gateway`）

| 実装 | 役割 |
|---|---|
| `openbd` | openBD から書誌と書影を取る（キー不要） |
| `rakuten` | 楽天ブックス書籍検索 API から書影を取る（アプリ ID とアクセスキーが要る）。通信エラーのメッセージから URL（キーを含む）を落とす |
| `catalog.NewChain` | 書誌は openBD だけから取り、openBD に書影が無いときだけ楽天の書影で補う。openBD に無い本は `ErrNotFound`。楽天の失敗は書影の補完を諦めるだけ。楽天のキーが無ければ openBD だけを使う |

HTTP クライアントのタイムアウトは 5 秒。

## 3. ユースケース（`usecase/book`）

| ユースケース | 流れ |
|---|---|
| 登録（command） | ISBN・感想・評価を VO で検証 → カタログを引く（該当なしは `ErrInvalid`、障害はそのままエラー）→ `New` → `Create` → 詳細を返す。検証を先に済ませ、不正な入力で外部カタログを呼ばない |
| 更新（command） | `FindByID` → 感想・評価を検証 → `ChangeReview` → カタログを取り直す（失敗・該当なしは今の書誌と書影のまま。失敗は warn ログ）→ `Update` → 詳細を返す |
| 削除（command） | `Delete`（無ければ `ErrNotFound`） |
| 取得（query） | `FindDetailByID` |
| 一覧（query） | `ListRange` を作って `FindList`。新しく登録した順 |
| カタログ確認（query） | ISBN を検証して `Lookup` |

## 4. 永続化（`infrastructure/postgres/book`）

- Repository と Query の実装。Persistence Model（`gorm` タグ付き）は Domain Model と別に持つ
- `Create` は一意制約 `uq_book_isbn` の違反を `ErrConflict` に変える（500 にしないため）
- `Update` は `id` と `version` の両方が一致する行だけを更新し、版を1つ進める
- 読み込み時は VO で検証し直す（壊れたデータを Domain に入れない）
- 一覧は `created_at DESC, id DESC`

## 5. HTTP（`presentation/http`）

- Handler は `BindValidate` → Command DTO を組み立てる → ユースケース → JSON。業務ロジックは持たない
- 書き込み系とカタログ確認には `RequireAdminToken` を掛ける。`Authorization: Bearer <ADMIN_TOKEN>` を定数時間で比較し、不一致・トークン未設定なら 401
- エラーの変換は `HTTPErrorHandler` の1か所:

| エラー | ステータス |
|---|---|
| 入力の検証（`BindValidate`） | 400（フィールドごとの `errors`） |
| `ErrInvalid` | 400 |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| それ以外 | 500（メッセージは隠す） |

## 6. 設定（`config`）

環境変数と既定値だけ（デプロイしないので環境ごとの切り替えは持たない）。既定値はローカルの docker compose の DB と、開発用の管理者トークン `local-admin-token`。楽天のキーは任意。

## 7. テストの分担

| 対象 | 方法 |
|---|---|
| VO・集約 | 単体テスト（境界値・トリム・制御文字） |
| ユースケース | 手書き Fake（Repository・Query・BookCatalog）で、検証エラー時に下位層を呼ばないこと、エラーの伝播、更新時のカタログ障害で更新が止まらないことを確かめる |
| Repository・Query | ローカル PostgreSQL での契約テスト |
| 外部カタログ | `httptest` のサーバに対する契約テスト |
| Handler・認証 | `httptest` + `common.NewEcho()` で 400・401・404・409 への変換を確かめる |
