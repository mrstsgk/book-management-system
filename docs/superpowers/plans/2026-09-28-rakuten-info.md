# 楽天由来の情報を持つ Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 楽天の書影を「画像 URL・商品ページ URL・取得日時」の3つ組で持ち、期限切れ（取得から89日）の楽天の書影を API が返さないようにする。

**Architecture:** `domain/book.Cover` に楽天専用のコンストラクタと期限判定を足し、楽天のゲートウェイが `itemUrl` と取得日時（注入した時計）で書影を作る。永続化はマイグレーション `000005` の2列で持ち、参照側（`postgres/book.query`）は期限切れの楽天の書影を null にして返す。

**Tech Stack:** Go / Echo / GORM / PostgreSQL / golang-migrate / swag / Orval

**Spec:** `docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md`（この計画は PR 1 = §2〜§5 だけ）

**マージの順序:** この PR は期限切れの楽天の書影を API が返さないようにするだけで、DB からは消さない。期限前の取り直しと期限切れの消去は同じ spec の PR 2（§6、ブランチ `feat/rakuten-refresh`。この PR の上に積む）が担う。PR 1 → PR 2 の順に続けてマージし、その間でリリースしない（このアプリはデプロイしないので、ローカルの develop で2本が揃うまでの間だけの状態）。

## Global Constraints

- 保持期限は取得から89日、取り直しの開始は取得から82日（spec §1）。定数は `domain/book` に1か所だけ置く
- 楽天の書影は画像 URL・商品ページ URL・取得日時が揃ったときだけ作れる。`NewCover` は openBD 専用にする
- 時刻はゲートウェイ・参照側とも注入した `func() time.Time` から取る（テストで固定するため）。ドメインは時計を持たない
- `RakutenDisabled`・`DisableRakuten`・`DropExpiredCover`・`rakuten_disabled` 列は §7/§6 の PR で足す（この PR では作らない。`rakuten_disabled` は `000006` にする）
- `backend/migrations/*.sql` を変えたら `docs/db/backend-schema.{json,md}` と `COMMENT ON COLUMN` を同じ変更に含める
- 生成物（swagger・`frontend/web/src/api/generated/`）は手編集しない。再生成してコミットしてからドリフト検査する
- 新しいテストの ISBN は `978000000340`〜`978000000359` + チェックディジット、テスト用の名前の接頭辞は `rakuten-test-`

## Review Focus

1. 取得からちょうど89日の楽天の書影は期限切れ（返さない）、88日台は返す。境界の向き（`<` と `<=`）を取り違えると規約違反か早すぎる消去になる。Task 1（VO）と Task 5（参照側）で固定する
2. openBD の書影は何年前のものでも返す（期限は楽天だけ）。Task 1・Task 5 で固定する
3. 楽天が `itemUrl` を返さない本には楽天の書影を付けない（商品ページへリンクできない書影は使えない）。Task 2 で固定する
4. 楽天の書影を保存して読み直すと、商品ページ・取得日時がそのまま戻る（タイムゾーンで取得日時がずれない）。Task 4 で固定する
5. マイグレーション後、`cover_source = 'rakuten'` なのに商品ページ・取得日時が欠けた行は作れない（CHECK 制約）。Task 3 で固定する

---

### Task 1: `Cover` に楽天の書影を足す

**Files:**
- Modify: `backend/internal/domain/book/cover.go`、`cover_test.go`

**Interfaces:**
- Produces:
  - `const RakutenRetention = 90 * 24 * time.Hour`、`const RakutenRefreshAfter = 83 * 24 * time.Hour`
  - `func RakutenExpired(fetchedAt, now time.Time) bool`（`!now.Before(fetchedAt.Add(RakutenRetention))`。参照側と VO が同じ規則を使うため）
  - `func NewRakutenCover(imageURL, productURL string, fetchedAt time.Time) (Cover, error)`
  - `func (c Cover) ProductURL() string`・`FetchedAt() time.Time`・`IsExpired(now time.Time) bool`・`NeedsRefresh(now time.Time) bool`

- [ ] **Step 1: 失敗するテストを書く**（`TestNewRakutenCover`・`TestCover_IsExpired`・`TestCover_NeedsRefresh`）
  - 有効な3つ組、商品ページが空・http・2049文字、画像が不正、取得日時がゼロ値はエラー
  - `IsExpired`: 楽天で取得から `89日-1ns` は false、`89日` ちょうどは true、openBD は `NewCover` で作ったものが10年後でも false
  - `NeedsRefresh`: `82日-1ns` は false、`82日` は true、openBD は false
- [ ] **Step 2: 失敗を確かめる** — Run: `cd backend && go test ./internal/domain/book/ -run 'Rakuten|IsExpired|NeedsRefresh'` / Expected: FAIL（`undefined: book.NewRakutenCover`）
- [ ] **Step 3: 実装** — URL の検証を `validCoverURL(raw string) bool` に切り出し、`NewCover` と `NewRakutenCover` の両方から使う。`Cover` に `productURL string`・`fetchedAt time.Time` を足す
- [ ] **Step 4: 通ることを見る** — Run: `cd backend && go test ./internal/domain/book/` / Expected: PASS
- [ ] **Step 5: コミット** — `feat: 楽天の書影を商品ページと取得日時つきで持てるようにする`

### Task 2: 楽天のゲートウェイが商品ページと取得日時で書影を作る

**Files:**
- Modify: `backend/internal/infrastructure/gateway/rakuten/catalog.go`、`catalog_test.go`
- Modify: `backend/internal/domain/book/cover.go`、`cover_test.go`（`NewCover` が楽天を拒む）
- Modify: `backend/internal/infrastructure/gateway/catalog/chain_test.go`、`backend/internal/presentation/http/catalog/handler_test.go`（楽天の書影の作り方を `NewRakutenCover` に）
- Modify: `backend/cmd/api/main.go`（`rakuten.NewCatalog(..., time.Now)`）

**Interfaces:**
- Consumes: Task 1 の `NewRakutenCover`
- Produces: `func NewCatalog(baseURL, applicationID, accessKey string, client *http.Client, now func() time.Time) domainbook.BookCatalog`

- [ ] **Step 1: 失敗するテストを書く**
  - 応答の `itemUrl` を商品ページ、注入した時刻を取得日時として書影が作られる
  - `itemUrl` が空なら書影なし（画像があっても）
  - `NewCover(url, CoverSourceRakuten)` は ErrInvalid（`cover_test.go` の「楽天の書影は有効」行を置き換える）
- [ ] **Step 2: 失敗を確かめる** — Run: `cd backend && go test ./internal/infrastructure/gateway/rakuten/ ./internal/domain/book/` / Expected: FAIL（`too many arguments in call to rakuten.NewCatalog` など）
- [ ] **Step 3: 実装** — `response.Items` に `ItemURL string \`json:"itemUrl"\`` を足し、`LargeImageURL != "" && ItemURL != ""` のときだけ `NewRakutenCover(it.LargeImageURL, it.ItemURL, c.now())`。`NewCover` は openBD 以外を拒む。テストの楽天の書影を `NewRakutenCover` に直す。`main.go` の `newCatalog` で `time.Now` を渡す
- [ ] **Step 4: 通ることを見る** — Run: `cd backend && go build ./... && go test ./internal/... ./cmd/...` / Expected: PASS
- [ ] **Step 5: コミット** — `feat: 楽天の書影を商品ページと取得日時つきで取得する`

### Task 3: マイグレーション `000005`

**Files:**
- Create: `backend/migrations/000005_add_book_rakuten_cover_info.up.sql`・`.down.sql`
- Modify: `docs/db/backend-schema.json`・`docs/db/backend-schema.md`

```sql
ALTER TABLE book ADD COLUMN cover_product_url VARCHAR(2048);
ALTER TABLE book ADD COLUMN cover_fetched_at TIMESTAMPTZ;

-- 既存の楽天の書影は商品ページも取得日時も分からないので外す（本を更新すれば取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'rakuten';

ALTER TABLE book ADD CONSTRAINT ck_book_rakuten_cover CHECK (
    (cover_source = 'rakuten' AND cover_product_url IS NOT NULL AND cover_fetched_at IS NOT NULL)
    OR (cover_source IS DISTINCT FROM 'rakuten' AND cover_product_url IS NULL AND cover_fetched_at IS NULL)
);

COMMENT ON COLUMN book.cover_product_url IS '楽天の商品ページの URL（楽天の書影のときだけ。楽天の規約上、書影と一緒にリンクする）。それ以外は NULL';
COMMENT ON COLUMN book.cover_fetched_at IS '楽天の書影・商品ページを取得した日時（保持期限89日の起点）。楽天の書影のときだけ。それ以外は NULL';
```

down: 制約・2列を落とす。

- [ ] **Step 1: 当てる** — Run: `cd backend && make migrate-up` / Expected: `5/u add_book_rakuten_cover_info`
- [ ] **Step 2: 制約を確かめる** — `psql` で `cover_source='rakuten'` かつ `cover_product_url` NULL の INSERT が `ck_book_rakuten_cover` で失敗し、openBD の書影に商品ページを付けた INSERT も失敗することを見る
- [ ] **Step 3: 往復** — `make migrate-down` → 直ちに `make migrate-up`（共有 DB なので1回だけ）
- [ ] **Step 4: DB 資料を直す**（JSON に2列と制約、MD の ER 図に2列）
- [ ] **Step 5: コミット** — `feat: 楽天の商品ページと取得日時の列を追加`

### Task 4: `postgres/book` の Repository が楽天の情報を保存・復元する

**Files:**
- Modify: `backend/internal/infrastructure/postgres/book/repository.go`、`repository_test.go`

- [ ] **Step 1: 失敗するテストを書く** — `TestRepository_RakutenCover`: 楽天の書影付きで作成して読むと画像・商品ページ・取得日時（`Equal` で比較）が戻る。更新で openBD の書影に差し替えると商品ページ・取得日時が消える
- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（読み直した書影の商品ページが空、または adapt が ErrInvalid）
- [ ] **Step 3: 実装** — `model` に `CoverProductURL *string`・`CoverFetchedAt *time.Time`。`toModel` は楽天のときだけ埋める。`Update` の map に2列。`adaptCover` は `source == rakuten` なら `NewRakutenCover`
- [ ] **Step 4: 通ることを見る** — Run: `cd backend && go test ./internal/infrastructure/postgres/book/ -count=1 -v`（SKIP なし）
- [ ] **Step 5: コミット** — `feat: 本のRepositoryで楽天の商品ページと取得日時を保存する`

### Task 5: 参照側が商品ページを返し、期限切れの楽天の書影を返さない

**Files:**
- Modify: `backend/internal/domain/book/book.go`（`BookDetail`・`BookListItem` に `CoverProductURL *string`）
- Modify: `backend/internal/infrastructure/postgres/book/query.go`、`query_test.go`

- [ ] **Step 1: 失敗するテストを書く** — `TestQuery_RakutenCover`（取得日時を直接書き込んで作る）: 取得から88日の楽天の書影は詳細・一覧とも画像・提供元・商品ページを返す、90日のものは3つとも nil、openBD の書影は古くても返し商品ページは nil
- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（`unknown field CoverProductURL` → 次に期限切れも返ってくる）
- [ ] **Step 3: 実装** — `query` に `now func() time.Time`（`NewQuery` の中で `time.Now`）。行から Read Model の書影3項目を作る関数 `coverFields(row, now)` を1つにし、詳細・一覧の両方から使う。一覧の `Select` に2列を足す。期限の判定は `domainbook.RakutenExpired`
- [ ] **Step 4: 通ることを見る** — Run: `cd backend && go test ./internal/... -count=1`
- [ ] **Step 5: コミット** — `feat: 期限切れの楽天の書影を返さず、商品ページを返す`

### Task 6: HTTP・OpenAPI・フロントの型

**Files:**
- Modify: `backend/internal/presentation/http/book/handler.go`、`handler_test.go`（`Response`・`ListItemResponse` に `coverProductUrl`）
- Modify: `backend/internal/presentation/http/catalog/handler.go`、`handler_test.go`（登録前の確認でも楽天の書影には商品ページを返す）
- Modify: `backend/api/docs/*`、`frontend/web/src/api/generated/*`（生成）
- Modify: `backend/architecture.md`（書影の節に商品ページと期限を1行）

- [ ] **Step 1: 失敗するテストを書く** — 詳細・一覧・カタログの確認の応答に `coverProductUrl` が載る
- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（`unknown field CoverProductURL`）
- [ ] **Step 3: 実装** — DTO に `CoverProductURL *string \`json:"coverProductUrl" example:"https://books.rakuten.co.jp/rb/15949390/"\``。変換で埋める
- [ ] **Step 4: 通ることを見る** — `go test ./... -count=1`・`golangci-lint run ./...`
- [ ] **Step 5: コミット** — `feat: 書影の応答に楽天の商品ページを追加`、`chore: 楽天の商品ページに合わせてAPIの型を再生成`、`chore: 楽天の書影の期限をドキュメントに反映`
- [ ] **Step 6: ドリフト確認** — `make swagger-check`・`pnpm gen:api:check`・`pnpm typecheck && pnpm lint && pnpm test && pnpm build`
