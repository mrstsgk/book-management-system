# 書影の取得元の見直し Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement each PR's task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 書影を Google Books → openBD → なし の順で取り、楽天を撤去する。

**Architecture:** PR-A（楽天の撤去）と PR-B（Google Books の Gateway 単体）を並行で作り、両方のマージ後に PR-C（Google Books を配線し、API・画面に出す）を作る。PR-B は Domain に触れない独立したパッケージにして、PR-A とぶつからないようにする。

**Tech Stack:** Go / Echo / GORM / golang-migrate / swag、React / Orval

> **変更（2026-09-30）:** PR-C（Google Books の配線、#93）は取りやめて閉じ、PR-B で入れた `gateway/googlebooks` は消した。Google Books も API キーが必要なため（[ADR](../../adr/2026-09-30-cover-from-openbd-only.md)）。PR-A（楽天の撤去）だけが有効。

**Spec:** `docs/superpowers/specs/2026-09-29-cover-sources-design.md`

## Global Constraints

- `CLAUDE.md` の「Before Changing Backend Go Code」の問いに答えてから変える。Repository の型と各メソッドには日本語の What コメント
- マイグレーションを変えたら `docs/db/backend-schema.{json,md}` を同じ PR で直し、SQL に `COMMENT ON` を書く
- 生成物（`backend/api/docs/`、`frontend/web/src/api/generated/`）は手編集しない。`make swagger` と `pnpm gen:api` で作り直す
- backend: `make fmt-check lint vuln test swagger-check`。frontend: `pnpm typecheck && pnpm lint && pnpm format:check && pnpm test && pnpm build && pnpm build-storybook && pnpm gen:api:check`
- 契約テストはローカル Postgres。ISBN は PR ごとに重ならない範囲（PR-A: 9780000006000〜6099、PR-C: 9780000007000〜7099）で、検査数字を正しく付け、実行ごとに一意にし、ID ごとに `t.Cleanup`
- ブランチ・コミット・PR は `docs/rules/git-*.md` に従う。push は `gh auth status` の有効アカウントが `mrstsgk` のときだけ（資格情報を取り出して回避しない）

## Review Focus

- 楽天の書影を持つ既存の行があるままマイグレーションを当てる → 書影が消え、制約違反で止まらない（000007 の up を、楽天の行を含むデータで確かめる）
- Google Books が 429・5xx・タイムアウト → 登録・更新は止まらず、openBD の書影（または書影なし）で保存される
- Google Books の `thumbnail` が `http://` → `https://` に直して保存する（Cover VO は https 以外を拒む）
- `GOOGLE_BOOKS_API_KEY` 未設定 → Google Books に一度も問い合わせない
- 起動時の書影の埋め直しで、書影が見つからない本 → 保存しない（版が上がらない）

---

### PR-A: 楽天の撤去（`refactor/remove-rakuten`）

**前提:** #85〜#87 のマージ後に develop から切る（#87 の詳細画面が楽天の表示を持つため）。

**Files（主なもの）:**
- Delete: `backend/internal/infrastructure/gateway/rakuten/`、`backend/internal/usecase/book/command/{refresh_rakuten_covers,disable_rakuten}{,_test}.go`
- Modify: `backend/internal/domain/book/{cover,book}.go`（楽天の VO・期限・`RakutenDisabled`・`DisableRakuten`・`DropExpiredCover`・`FindRakutenRefreshTargets` を消す）、`backend/internal/infrastructure/gateway/catalog/chain.go`（fallback を外して openBD だけを返す。PR-C で Google Books を足す）、`backend/internal/infrastructure/postgres/book/*`、`backend/internal/presentation/http/book/*`（`DELETE /rakuten`・`coverProductUrl`・`rakutenDisabled` を消す）、`backend/cmd/api/main.go`（取り直しの goroutine を消す）、`backend/config/config.go`
- Create: `backend/migrations/000007_remove_rakuten.{up,down}.sql`
- Modify（frontend）: 詳細画面の楽天の表示、`Footer`、生成物の再生成
- Modify（docs）: `docs/specifications.md` §1.1・§1.2・§2.2、`backend/architecture.md`、`backend/README.md`、`docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md` と `2026-09-29-public-screens-design.md` に「撤去」の注記、`docs/design/public-screens/detail.html` と PNG、既存 ADR `2026-09-28-rakuten-for-cover-only.md` の状態を「撤去」に
- Create（docs）: `docs/adr/2026-09-29-cover-from-google-books-and-openbd.md`、本計画と spec（main checkout にある未コミットのファイルをコピーしてコミットする）

- [ ] **Step 1:** Domain のテストから楽天の分を消し、残る振る舞い（openBD の Cover、`RefreshCatalog` が書誌と書影を差し替える）のテストが通ることを確かめる → 楽天の型と関数を消す
- [ ] **Step 2:** 000007 の up/down を書き、楽天の書影を持つ行を入れた一時 DB で up → down → up を確かめる（共有の開発 DB では試さない）。DB の資料を直す
- [ ] **Step 3:** Infrastructure・UseCase・Presentation・main から楽天を消し、`make swagger`。Handler のテストと結合テストから `DELETE /rakuten` を消す
- [ ] **Step 4:** frontend: `pnpm gen:api`、詳細とフッターの楽天の表示を消し、テストを直す（楽天の表示が出ないことではなく、残る表示を確かめる）
- [ ] **Step 5:** docs と ADR。全チェック、コミット（意味ごと）、push、PR

### PR-B: Google Books の Gateway（`feat/google-books-cover-gateway`、すぐ着手できる）

**Files:** Create `backend/internal/infrastructure/gateway/googlebooks/{client,client_test}.go` だけ（Domain・配線・config には触れない）

**Interfaces（PR-C が使う）:**
```go
package googlebooks

// Cover は Google Books が持つ書影。PR-C で domain の Cover に詰め替える。
type Cover struct {
	ImageURL string // https に直した imageLinks.thumbnail
	PageURL  string // volumeInfo.infoLink（その本の Google Books のページ）
}

// NewClient は baseURL（既定 https://www.googleapis.com）と API キーで作る。
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client

// FindCover は ISBN の書影を返す。該当なし・imageLinks なしは (nil, nil)。通信・HTTP の失敗は error。
func (c *Client) FindCover(ctx context.Context, isbn string) (*Cover, error)
```

- [ ] **Step 1:** `httptest.Server` で契約テストを先に書く: 正常（thumbnail が http → https、infoLink を PageURL に）／`totalItems: 0`／`imageLinks` なし／`infoLink` なし（書影なしとして扱う。Google の規約上リンクが要るため）／429・500（error）／壊れた JSON（error）／要求に `q=isbn:<ISBN>` と `key=<キー>` が付く／ctx のキャンセル
- [ ] **Step 2:** 実装して通す。全チェック、コミット、push、PR

### PR-C: Google Books の配線と表示（`feat/cover-from-google-books`、PR-A と PR-B のマージ後）

**Files（主なもの）:** `backend/internal/domain/book/cover.go`（`CoverSourceGoogleBooks`・`pageURL`・`NewGoogleBooksCover`）、`backend/migrations/000008_add_book_cover_page_url.{up,down}.sql`、`backend/internal/infrastructure/gateway/catalog/`（`New(openbd, googlebooks)`。Google Books の書影があれば差し替え、失敗は warn）、`backend/internal/infrastructure/postgres/book/*`、`backend/internal/usecase/book/command/fill_missing_covers{,_test}.go`（起動時に1回）、`backend/cmd/api/main.go`、`backend/config/config.go`（`GOOGLE_BOOKS_API_KEY`・`GOOGLE_BOOKS_BASE_URL`）、Presentation の `coverPageUrl`、frontend の一覧・詳細・フッター、画面イメージ、`backend/README.md`

- [ ] **Step 1:** Domain: `NewGoogleBooksCover(imageURL, pageURL)` のテスト（どちらかが不正・空ならエラー）→ 実装
- [ ] **Step 2:** catalog の組み合わせのテスト（Google Books あり → 差し替え／なし → openBD のまま／失敗 → openBD のまま＋warn／googlebooks nil → 問い合わせない／openBD が NotFound → そのまま NotFound）→ 実装
- [ ] **Step 3:** 000008 と永続化、契約テスト。DB の資料
- [ ] **Step 4:** `FillMissingCoversUsecase`（書影の無い本だけ、見つからなければ保存しない、1冊の失敗で止めない）のテスト → 実装 → main で起動時に1回
- [ ] **Step 5:** API の `coverPageUrl`、`make swagger`、frontend（詳細の「Powered by Google」と「Google Books で見る」、一覧のカードの外のリンク、フッター）、画面イメージと PNG
- [ ] **Step 6:** 全チェック、コミット、push、PR
