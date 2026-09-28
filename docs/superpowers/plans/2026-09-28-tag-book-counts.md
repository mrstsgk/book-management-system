# 分野別の集計 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 分野タグごとの冊数を返す `GET /api/tags/counts` を追加する。

**Architecture:** `domain/tag` に集計の Read Model と `Query.CountBooks` を足し、`postgres/tag` で `tag JOIN book_tag GROUP BY` の1クエリで返す。`usecase/tag/query` の薄いユースケースと、`presentation/http/tag` の認証なしのハンドラでつなぐ。

**Tech Stack:** Go / Echo / GORM / PostgreSQL / swag / Orval

**Spec:** `docs/superpowers/specs/2026-09-28-book-list-search-design.md` §2（§1 の検索・絞り込みは別の PR）

## Global Constraints

- 本が1冊も付いていないタグは含めない
- 冊数の多い順、同数ならタグ名順
- 認証不要
- 集計の Read Model は書き込み側の VO（`tag.Name`）を使わず `string` で持つ
- 契約テストのデータは ISBN `9780000003201`〜`9780000003294`、名前の接頭辞 `counts-test-` だけを使う（共有 DB に他のテストの行が残っていても通るように、自分が作った行だけを検証する）

## Review Focus

1. 同じ本にタグが複数付いていても、各タグの冊数がその本を1冊として数える（JOIN で他タグの行が混ざって冊数が増えない）。Task 1 の契約テストで固定する
2. 0冊のタグが結果に出ない（`LEFT JOIN` にしてしまうと冊数0で出る）。Task 1 で固定する
3. タグを削除した後の集計にそのタグが出ない（`ON DELETE CASCADE` 任せ）。Task 1 で固定する
4. `GET /api/tags/counts` が既存の `PUT/DELETE /api/tags/:id` とぶつからず、認証なしで取れる。Task 3 のハンドラテストで固定する
5. 冊数が同じときの並びがタグ名順で安定する。Task 1 で固定する

---

### Task 1: ドメインの Read Model と `postgres/tag` の集計

**Files:**
- Modify: `backend/internal/domain/tag/tag.go`
- Modify: `backend/internal/infrastructure/postgres/tag/query.go`、`query_test.go`
- Modify（Fake に新メソッドを足す）: `backend/internal/usecase/tag/query/list_test.go`、`backend/internal/usecase/book/command/register_test.go`、`backend/internal/usecase/book/command/tag_selection_test.go`

**Interfaces:**
- Produces: `tag.TagBookCount{ID ID; Name string; BookCount int}`、`tag.TagBookCounts{Items []*TagBookCount}`、`tag.Query.CountBooks(ctx) (*TagBookCounts, error)`

- [ ] **Step 1: 失敗するテストを書く**（`query_test.go`）

タグ a（2冊）・b1・b2（1冊ずつ）・c（0冊）、本1にタグ a と b1、本2にタグ a と b2 を付ける。自分の ID の行だけを取り出して検証する: a=2・b1=1・b2=1、c は出ない、並びは a → b1 → b2。タグ d（1冊）を作って削除すると出ない。

- [ ] **Step 2: 失敗を確かめる** — `go test ./internal/infrastructure/postgres/tag/ -run TestQuery_CountBooks` → ビルドエラー（`CountBooks` 未定義）

- [ ] **Step 3: 最小の実装**

`tag.go` に Read Model と `Query.CountBooks` を足し、`query.go` で:

```go
func (q *query) CountBooks(ctx context.Context) (*domaintag.TagBookCounts, error) {
	var rows []struct {
		ID        int64
		Name      string
		BookCount int
	}
	err := q.db.WithContext(ctx).Table("tag").
		Select("tag.id AS id, tag.name AS name, COUNT(*) AS book_count").
		Joins("JOIN book_tag ON book_tag.tag_id = tag.id").
		Group("tag.id, tag.name").
		Order("book_count DESC, tag.name").
		Find(&rows).Error
	...
}
```

既存の Fake 3つに `CountBooks` を足す（`nil, nil` を返す）。

- [ ] **Step 4: 通ることを見る** — `go test ./internal/infrastructure/postgres/tag/ -v`（SKIP なし）と `go build ./...`

- [ ] **Step 5: コミット** — `feat: 分野タグごとの冊数を集計するQueryを追加`

### Task 2: ユースケース

**Files:**
- Create: `backend/internal/usecase/tag/query/count_books.go`、`count_books_test.go`

**Interfaces:**
- Consumes: `tag.Query.CountBooks`
- Produces: `query.CountBooksUsecase{Execute(ctx) (*tag.TagBookCounts, error)}`、`query.CountBooksUsecaseImpl{Tags tag.Query}`

- [ ] **Step 1: 失敗するテストを書く** — Fake の結果をそのまま返す、Query のエラーをそのまま返す
- [ ] **Step 2: 失敗を確かめる** — ビルドエラー（`CountBooksUsecaseImpl` 未定義）
- [ ] **Step 3: 実装** — `return u.Tags.CountBooks(ctx)`
- [ ] **Step 4: 通ることを見る** — `go test ./internal/usecase/tag/query/ -v`
- [ ] **Step 5: コミット** — `feat: 分野タグごとの冊数を返すユースケースを追加`

### Task 3: HTTP・配線・OpenAPI・ドキュメント

**Files:**
- Modify: `backend/internal/presentation/http/tag/handler.go`、`handler_test.go`
- Modify: `backend/cmd/api/main.go`、`main_test.go`
- Modify: `backend/api/docs/*`、`frontend/web/src/api/generated/*`（生成）
- Modify: `backend/architecture.md`

**Interfaces:**
- Consumes: `query.CountBooksUsecase`
- Produces: `GET /api/tags/counts` → `{"items":[{"id","name","bookCount"}]}`（`TagBookCountListResponse`）

- [ ] **Step 1: 失敗するテストを書く**（`handler_test.go`）— トークン無しで200と本文、0件で `"items":[]`、ユースケースのエラーが500、`PUT /api/tags/counts` は `:id` のパースで400（ぶつからない）。`main_test.go` のルート一覧に `GET /api/tags/counts` を足す
- [ ] **Step 2: 失敗を確かめる** — ビルドエラー（`CountBooksUC` 未定義）
- [ ] **Step 3: 実装** — `Handler.CountBooksUC`、`g.GET("/counts", h.CountBooks)`（swag コメント付き）、`main.go` で `&tagqry.CountBooksUsecaseImpl{Tags: tagQuery}` を渡す
- [ ] **Step 4: 通ることを見る** — `go test ./... -count=1`、`golangci-lint run ./...`
- [ ] **Step 5: コミット** — `feat: 分野タグごとの冊数のAPIを追加`
- [ ] **Step 6: 再生成** — `make swagger`、`pnpm gen:api` → コミット `chore: 分野別の集計に合わせてAPIの型を再生成`
- [ ] **Step 7: ドキュメント** — `backend/architecture.md` の API 表に1行 → コミット `chore: 分野別の集計APIをドキュメントに反映`
- [ ] **Step 8: ドリフト確認** — `make swagger-check`、`pnpm gen:api:check`、`pnpm typecheck && pnpm lint && pnpm test && pnpm build`
