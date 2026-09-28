# 楽天からの削除指示への対応 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 楽天から削除の指示を受けた本の楽天由来の情報を管理画面から消せるようにし、以後その本に楽天の書影が付かないようにする。

**Architecture:** `Book` に `RakutenDisabled` を持たせ、`DisableRakuten()` で楽天の書影を外して立てる。`RefreshCatalog` は立っている本に楽天の書影を付けない。`DisableRakutenUsecase` が読んだ直後の version で保存し、競合したら1回だけ読み直す。`DELETE /api/books/{id}/rakuten` から呼ぶ。

**Tech Stack:** Go / Echo / GORM / PostgreSQL / golang-migrate / swag / Orval

**Spec:** `docs/superpowers/specs/2026-09-28-rakuten-expiry-design.md` §2（`RakutenDisabled`・`DisableRakuten`・`RefreshCatalog`）・§7。§6（取り直し・`DropExpiredCover`）は別 PR。

## Global Constraints

- PR #75（`feat/rakuten-expiry`）の上に積む。マイグレーションは `000006`
- `rakuten_disabled BOOLEAN NOT NULL DEFAULT FALSE`、`COMMENT ON COLUMN` と `docs/db/backend-schema.{json,md}` を同じ変更で更新する
- `DELETE /api/books/{id}/rakuten`: 認証必要・204・存在しない本は 404・`version` は受け取らない
- 楽天の書影が付いていない本・既に無効化した本でも 204
- Repository の IF とメソッドには1行の日本語 What コメント

## Review Focus

1. 無効化した本を更新（`PUT`）しても、外部カタログが楽天の書影を返したときに書影が付かない — Task 1 の `RefreshCatalog` と Task 3 の `UpdateUsecase` 経由テスト
2. 無効化した本でも openBD の書影は付く — Task 1
3. 無効化の保存が他の更新と競合したとき、1回だけ読み直して成功する／2回続けて競合したら 409 — Task 3
4. 既に無効化済みで楽天の書影も無い本に `DELETE` しても保存せずに 204 — Task 3
5. 無効化したことが再読み込み後も残る（Repository の往復）— Task 2

---

### Task 1: ドメイン（`RakutenDisabled`・`DisableRakuten`・`RefreshCatalog`）

**Files:** Modify `backend/internal/domain/book/book.go`、`book_test.go`

- [ ] Step 1: 失敗するテストを書く（`book_test.go`）
  - `DisableRakuten` は楽天の書影を外し `RakutenDisabled` を立てる
  - `DisableRakuten` は openBD の書影を残す
  - 無効化後の `RefreshCatalog` に楽天の書影を渡すと書影なし、openBD の書影なら付く
  - 無効化していない本の `RefreshCatalog` は楽天の書影を付ける（既存の挙動）
- [ ] Step 2: `go test ./internal/domain/book/` → FAIL（`RakutenDisabled`・`DisableRakuten` が無い）
- [ ] Step 3: 実装

```go
// RakutenDisabled は楽天から削除の指示を受けた本。以後、この本には楽天の書影を付けない。
RakutenDisabled bool

// DisableRakuten は楽天由来の情報（楽天の書影）を外し、以後付けないようにする。
func (b *Book) DisableRakuten() {
	if b.Cover != nil && b.Cover.Source() == CoverSourceRakuten {
		b.Cover = nil
	}
	b.RakutenDisabled = true
}

// RefreshCatalog: RakutenDisabled の本に楽天の書影が来たら書影なしにする
```

- [ ] Step 4: PASS
- [ ] Step 5: `git commit -m "feat: 楽天から削除の指示を受けた本に楽天の書影を付けないようにする"`

### Task 2: マイグレーション・Repository・Query

**Files:** Create `backend/migrations/000006_add_book_rakuten_disabled.{up,down}.sql`。Modify `postgres/book/repository.go`・`repository_test.go`・`query.go`・`query_test.go`・`domain/book/book.go`（`BookDetail.RakutenDisabled`）・`docs/db/backend-schema.{json,md}`

- [ ] Step 1: マイグレーション

```sql
ALTER TABLE book ADD COLUMN rakuten_disabled BOOLEAN NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN book.rakuten_disabled IS '楽天から削除の指示を受けた本なら TRUE。以後この本には楽天の書影を付けない';
```

down: `ALTER TABLE book DROP COLUMN IF EXISTS rakuten_disabled;`

- [ ] Step 2: `make migrate-up`、往復確認1回（down → 直ちに up）
- [ ] Step 3: 失敗する契約テスト（`TestRepository_RakutenDisabled`: 無効化して Update → FindByID で `RakutenDisabled` が true、楽天の書影が消えている／`TestQuery_RakutenDisabled`: 詳細が `RakutenDisabled` を返す）。ISBN は `978000000370`〜`379`＋チェックディジット
- [ ] Step 4: FAIL を確かめる
- [ ] Step 5: `model.RakutenDisabled bool \`gorm:"column:rakuten_disabled;not null"\``、`toModel`・`adapt`・`Update` の列、`BookDetail.RakutenDisabled`、Query の詳細で詰める
- [ ] Step 6: PASS（SKIP なし）
- [ ] Step 7: DB 資料を更新し `git commit -m "feat: 楽天から削除の指示を受けたことを保存する"`

### Task 3: ユースケース `DisableRakutenUsecase`

**Files:** Create `backend/internal/usecase/book/command/disable_rakuten.go`・`disable_rakuten_test.go`

- [ ] Step 1: 失敗するテスト（Fake）
  - 楽天の書影を外して `RakutenDisabled` を立てて保存する（読んだ version のまま渡す）
  - 1回目の保存が `ErrConflict` なら読み直してやり直し、成功する
  - 2回続けて `ErrConflict` なら `ErrConflict` を返す
  - 存在しない本は `ErrNotFound`（保存しない）
  - 既に無効化済みで楽天の書影も無い本は保存せずに成功する
  - 無効化した本を `UpdateUsecase` で更新し、カタログが楽天の書影を返しても書影が付かない（Review Focus 1）
- [ ] Step 2: FAIL
- [ ] Step 3: 実装

```go
// DisableRakutenUsecaseImpl.Execute は、読んだ直後の version で保存し、競合したら1回だけ読み直してやり直す
// （削除の指示には画面で最後に読んだ内容に関係なく従う必要があるため version を受け取らない）。
func (u *DisableRakutenUsecaseImpl) Execute(ctx context.Context, id int64) error {
	for attempt := 0; ; attempt++ {
		b, err := u.Books.FindByID(ctx, book.ID(id))
		if err != nil {
			return err
		}
		if b.RakutenDisabled && (b.Cover == nil || b.Cover.Source() != book.CoverSourceRakuten) {
			return nil
		}
		b.DisableRakuten()
		err = u.Books.Update(ctx, b)
		if errors.Is(err, common.ErrConflict) && attempt == 0 {
			continue
		}
		return err
	}
}
```

- [ ] Step 4: PASS
- [ ] Step 5: `git commit -m "feat: 楽天由来の情報を消すユースケースを追加"`

### Task 4: HTTP・配線・生成物・ドキュメント

**Files:** Modify `presentation/http/book/handler.go`・`handler_test.go`、`cmd/api/main.go`・`main_test.go`、`backend/api/docs/*`、`frontend/web/src/api/generated/*`、`backend/architecture.md`

- [ ] Step 1: 失敗するテスト（Handler: トークンありで 204・usecase に path の ID が渡る、トークン無しで 401、存在しない本は 404、詳細の応答に `rakutenDisabled`。main_test のルート一覧に `DELETE /api/books/:id/rakuten`）
- [ ] Step 2: FAIL
- [ ] Step 3: `Handler.DisableRakutenUC`、`g.DELETE("/:id/rakuten", h.DisableRakuten, h.AdminOnly)`（swag コメント付き）、`Response.RakutenDisabled bool \`json:"rakutenDisabled"\``、`main.go` で配線
- [ ] Step 4: PASS、`go test ./... -count=1`・`golangci-lint run ./...`
- [ ] Step 5: `git commit -m "feat: 楽天由来の情報を消すAPIを追加"`
- [ ] Step 6: `make swagger`・`pnpm gen:api` → `git commit -m "chore: 楽天由来の情報を消すAPIに合わせて型を再生成"` → `make swagger-check`・`pnpm gen:api:check`・`pnpm typecheck && pnpm lint && pnpm test && pnpm build`
- [ ] Step 7: `backend/architecture.md` の API 表に1行 → `git commit -m "chore: 楽天由来の情報を消すAPIをドキュメントに反映"`
