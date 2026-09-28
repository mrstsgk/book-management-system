# 楽天由来の情報の取り直しと消去 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 楽天の書影を期限の7日前から取り直し、取り直せないまま期限を過ぎたら外す（起動時・1日1回・本の更新時）。

**Architecture:** 期限の判定は既存の `Cover.NeedsRefresh`/`IsExpired`（PR #75）だけを使う。取り直しは `usecase/book/command` の `RefreshRakutenCoversUsecase` が、`book.Repository.FindRakutenRefreshTargets` で対象を読み、既存の `BookCatalog` で取り直して `Update` する。起動時と24時間ごとの呼び出しは `cmd/api` の goroutine。

**Tech Stack:** Go / GORM / PostgreSQL（既存）

**Spec:** `docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md` §6（PR 2）。§7（`RakutenDisabled`・削除 API）は別 PR。

## Global Constraints

- 保持期限 89日、取り直し開始 82日（`book.RakutenRetention`・`book.RakutenRefreshAfter`、PR #75 で定義済み）。二つ目の期限規則を作らない
- 時刻はユースケースに `Now func() time.Time` で注入する。ドメインは時計を持たない
- 1冊の失敗で残りを止めない。失敗は warn ログ
- 起動を取り直しで待たせない
- Repository の IF と実装のメソッドには日本語1行の What コメント
- テスト用 ISBN は 9780000003607〜9780000003690、テスト用接頭辞は `refresh-test-`

## Review Focus

1. 取り直しが `ErrConflict`（同時に自分が更新した）を返した本は飛ばし、残りは続ける — Task 3 のテスト
2. 取り直せず期限前の本は保存しない（version を無駄に進めない） — Task 3 のテスト
3. openBD の書影は古くても対象にならない — Task 2 の契約テスト
4. 取り直し開始（82日）ちょうどの本が対象に入り、それ未満は入らない — Task 2 の契約テスト
5. shutdown（context のキャンセル）で定期実行が止まる — Task 5 のテスト

---

### Task 1: `Book.DropExpiredCover`

**Files:** Modify `backend/internal/domain/book/book.go`、`book_test.go`

**Interfaces:** Produces `func (b *Book) DropExpiredCover(now time.Time) bool`（外したら true）

- [ ] **Step 1: 失敗するテスト**（`book_test.go`）: 楽天の書影が保持期限ちょうどなら外して true、その1秒前なら残して false、openBD の書影は何年たっても残して false、書影なしは false
- [ ] **Step 2:** `go test ./internal/domain/book/ -run TestBook_DropExpiredCover` → FAIL（undefined）
- [ ] **Step 3: 実装**

```go
// DropExpiredCover は楽天の書影が保持期限を過ぎていれば外し、外したかを返す。
func (b *Book) DropExpiredCover(now time.Time) bool {
	if b.Cover == nil || !b.Cover.IsExpired(now) {
		return false
	}
	b.Cover = nil
	return true
}
```

- [ ] **Step 4:** PASS
- [ ] **Step 5:** `git commit -m "feat: 保持期限を過ぎた楽天の書影を本から外せるようにする"`

### Task 2: `Repository.FindRakutenRefreshTargets`

**Files:** Modify `backend/internal/domain/book/book.go`（IF）、`backend/internal/infrastructure/postgres/book/repository.go`、`repository_test.go`、`backend/internal/usecase/book/command/register_test.go`（Fake）

**Interfaces:** Produces `FindRakutenRefreshTargets(ctx context.Context, fetchedBefore time.Time) ([]*Book, error)` — 楽天の書影で `cover_fetched_at <= fetchedBefore` の本を ID 順に返す（`NeedsRefresh` の `!now.Before(fetchedAt+83d)` と同じ境界）

- [ ] **Step 1: 失敗する契約テスト**: 基準時刻 `now` に対し、取り直し開始ちょうど前（9780000003607）・保持期限超過（…3614）の楽天の書影は返り、取り直し開始の1秒後（…3621）の楽天の書影・古い openBD の書影（…3638）・書影なし（…3645）は返らない。返った本はタグも読めている（…3607 にタグ1つ）。共有 DB に他の行があるので、自分の ID だけを見て判定する
- [ ] **Step 2:** FAIL（undefined）
- [ ] **Step 3: 実装**: IF に追加。postgres は `Where("cover_source = ? AND cover_fetched_at <= ?", "rakuten", fetchedBefore).Order("id")` で行を取り、1冊ずつ `findBookTagIDs`＋`adapt`（対象は多くて数冊なので1冊ずつで足りる）。Fake にも空実装
- [ ] **Step 4:** PASS（SKIP でないこと）
- [ ] **Step 5:** `git commit -m "feat: 取り直す時期の楽天の書影を持つ本を読めるようにする"`

### Task 3: `RefreshRakutenCoversUsecase`

**Files:** Create `backend/internal/usecase/book/command/refresh_rakuten_covers.go`、`refresh_rakuten_covers_test.go`

**Interfaces:** Consumes Task 1・2。Produces

```go
type RefreshRakutenCoversUsecase interface {
	Execute(ctx context.Context) error
}
type RefreshRakutenCoversUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Now     func() time.Time
}
```

- [ ] **Step 1: 失敗するテスト**（Fake の Repository/Catalog）:
  - `FindRakutenRefreshTargets` に `now - RakutenRefreshAfter` が渡る
  - 取り直せたら新しい書影で `Update` する
  - 取り直せず（障害・該当なし）期限切れなら書影を外して `Update` する
  - 取り直せず期限前なら `Update` しない
  - 1冊目の `Update` が `ErrConflict` でも2冊目を処理する
  - 1冊目の `Update` が他のエラーでも2冊目を処理する
  - `FindRakutenRefreshTargets` のエラーは返す
- [ ] **Step 2:** FAIL
- [ ] **Step 3: 実装**（複雑度を10以下に保つため、1冊分を `refresh(ctx, b, now)` に分ける）
- [ ] **Step 4:** PASS
- [ ] **Step 5:** `git commit -m "feat: 取り直す時期の楽天の書影を取り直し、期限切れなら外すユースケースを追加"`

### Task 4: 本の更新時に期限切れを外す

**Files:** Modify `backend/internal/usecase/book/command/update.go`、`update_test.go`

**Interfaces:** `UpdateUsecaseImpl` に `Now func() time.Time` を足す

- [ ] **Step 1: 失敗するテスト**: 楽天の書影が期限切れの本を、カタログ障害のまま更新すると書影が外れて保存される。期限前なら残る。既存テストの `UpdateUsecaseImpl` には固定の `Now` を渡す
- [ ] **Step 2:** FAIL
- [ ] **Step 3: 実装**: `refreshCatalog` の失敗時に `b.DropExpiredCover(u.Now())`
- [ ] **Step 4:** PASS
- [ ] **Step 5:** `git commit -m "feat: 本の更新で書影を取り直せず期限切れなら楽天の書影を外す"`

### Task 5: 起動時と1日1回の呼び出し

**Files:** Modify `backend/cmd/api/main.go`、`main_test.go`

**Interfaces:** Produces `func startRakutenRefresh(ctx context.Context, uc bookcmd.RefreshRakutenCoversUsecase, interval time.Duration)` — goroutine を起動してすぐ返す。起動直後に1回、その後 interval ごと。ctx が終われば止まる。エラーは warn ログ

- [ ] **Step 1: 失敗するテスト**（`main_test.go`）: Execute がブロックする Fake でも `startRakutenRefresh` がすぐ返る（起動を待たせない）。interval を短くすると2回以上呼ばれる。ctx をキャンセルするとそれ以上呼ばれない
- [ ] **Step 2:** FAIL
- [ ] **Step 3: 実装**: `run()` で `ctx, cancel := context.WithCancel(context.Background())`・`defer cancel()`（`run` は Serve のグレースフルシャットダウン後に返るので、そこで止まる）。`startRakutenRefresh(ctx, &bookcmd.RefreshRakutenCoversUsecaseImpl{...Now: time.Now}, 24*time.Hour)` を `Serve` の直前に呼ぶ。`registerRoutes` の `UpdateUsecaseImpl` に `Now: time.Now`
- [ ] **Step 4:** PASS
- [ ] **Step 5:** `git commit -m "feat: 起動時と1日1回、楽天の書影を取り直す"`

### Task 6: 仕上げ

- [ ] `go build ./... && go vet ./... && go test ./... -count=1`（SKIP なし）、`golangci-lint run ./...` 0 issues、`make swagger-check`（API 変更なし）
- [ ] `backend/architecture.md` の「書誌・書影」の行に「楽天の書影は起動時・1日1回・更新時に取り直し、保持期限（89日）で外す」を足す → `git commit -m "chore: 楽天の書影の取り直しをドキュメントに反映"`
