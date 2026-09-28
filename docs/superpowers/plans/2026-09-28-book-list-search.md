# 一覧の検索・絞り込み Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `GET /api/books` にキーワード（書名・著者の部分一致）と分野タグの絞り込みを足す。

**Architecture:** 参照側の条件 VO `book.ListCondition` を足し、`book.Query.FindList` に渡す。ユースケースは入力構造体 `ListInput` を受け取る。Postgres 実装は `ILIKE ... ESCAPE` と `EXISTS (book_tag)` で絞り、総件数にも同じ条件を掛ける。

**Tech Stack:** Go / Echo / GORM / PostgreSQL / swag / Orval

**Spec:** `docs/superpowers/specs/2026-09-28-book-list-search-design.md`（§1 のみ。§2 分野別の集計は別 PR）

## Global Constraints

- `q` は前後の空白を除いて100文字まで、制御文字は不可。空・空白だけは検索しない
- `tagId` は1以上。存在しない `tagId` は200・0件
- 書名は `COALESCE(title_override, title)` で検索する（上書き前の書名では当たらない）
- `%`・`_`・`\` は文字として扱う
- 大文字小文字を区別しない（`ILIKE`）
- `total` は条件に合う件数
- インデックスは足さない
- 参照側の VO は書き込み側の VO を使わない
- 契約テストの ISBN は `9780000003003`〜`9780000003096`（978000000300〜309 + チェックディジット）、書名の接頭辞は `search-test-`

## Review Focus

1. 上書きした書名で当たり、上書き前の書名では当たらない → Task 2 の契約テスト
2. `%`・`_` を含むキーワードが全件に当たらない → Task 2 の契約テスト
3. タグが2つ付いた本がタグ絞り込みで重複しない（JOIN ではなく EXISTS） → Task 2 の契約テスト
4. `total` が絞り込み後の件数になる → Task 2 の契約テスト
5. `tagId=0`・`q` 101文字・`tagId=abc` が400 → Task 4 の Handler テスト

---

### Task 1: `book.ListCondition`（参照側の条件 VO）

**Files:**
- Create: `backend/internal/domain/book/list_condition.go`
- Test: `backend/internal/domain/book/list_condition_test.go`

**Interfaces:**
- Produces: `book.NewListCondition(keyword string, tagID *tag.ID) (ListCondition, error)`、`(ListCondition) Keyword() string`、`(ListCondition) TagID() (tag.ID, bool)`

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestNewListCondition(t *testing.T) {
	t.Parallel()
	id := func(v int64) *tag.ID { x := tag.ID(v); return &x }
	tests := []struct {
		name        string
		keyword     string
		tagID       *tag.ID
		wantKeyword string
		wantTag     tag.ID
		wantHasTag  bool
		wantErr     bool
	}{
		{name: "ゼロ値は絞り込まない", keyword: ""},
		{name: "空白だけは検索しない", keyword: " 　"},
		{name: "前後の空白を除く", keyword: " 設計 ", wantKeyword: "設計"},
		{name: "100文字ちょうどは有効", keyword: strings.Repeat("あ", 100), wantKeyword: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", keyword: strings.Repeat("あ", 101), wantErr: true},
		{name: "制御文字はエラー", keyword: "設計\n入門", wantErr: true},
		{name: "タグを指定できる", tagID: id(3), wantTag: 3, wantHasTag: true},
		{name: "タグID 0はエラー", tagID: id(0), wantErr: true},
	}
	// ...各ケースで Keyword()・TagID() と ErrInvalid を確かめる
}
```

- [ ] **Step 2: 失敗を確かめる** — Run: `cd backend && go test ./internal/domain/book/ -run TestNewListCondition` / Expected: FAIL（undefined: book.NewListCondition）
- [ ] **Step 3: 実装**

```go
const keywordMaxLength = 100

type ListCondition struct {
	keyword string
	tagID   *tag.ID
}

func NewListCondition(keyword string, tagID *tag.ID) (ListCondition, error) {
	if strings.IndexFunc(keyword, unicode.IsControl) >= 0 { /* ErrInvalid */ }
	v := strings.TrimSpace(keyword)
	if utf8.RuneCountInString(v) > keywordMaxLength { /* ErrInvalid */ }
	if tagID != nil && *tagID < 1 { /* ErrInvalid */ }
	// tagID はコピーして持つ
}
```

- [ ] **Step 4: 通ることを見る** — Expected: PASS
- [ ] **Step 5: コミット** — `feat: 一覧の検索・絞り込み条件のVOを追加`

### Task 2: `book.Query.FindList` に条件を渡す（Postgres）

**Files:**
- Modify: `backend/internal/domain/book/book.go`（`Query.FindList` のシグネチャ）
- Modify: `backend/internal/infrastructure/postgres/book/query.go`、`query_test.go`
- Modify: Fake の `FindList`（`usecase/book/query/get_test.go`、`usecase/book/command/register_test.go`）

**Interfaces:**
- Consumes: Task 1 の `ListCondition`
- Produces: `FindList(ctx context.Context, c ListCondition, r common.ListRange) (*BookList, error)`

- [ ] **Step 1: 失敗するテストを書く** — `TestQuery_FindList_Condition`（書名の部分一致・大文字小文字の無視・上書き書名で当たり元の書名で当たらない・著者の部分一致・`%`/`_` を文字として扱う・タグ絞り込み・タグ2つの本が重複しない・キーワードとタグの AND・`total` が絞り込み後・存在しないタグは0件）。書名は `search-test-` で始め、キーワードにも `search-test-` を含めて他のテストの行と混ざらないようにする。著者は作成後に `UPDATE book SET authors = ?` で変える
- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（コンパイルエラー: FindList の引数）
- [ ] **Step 3: 実装** — `applyCondition(tx, c)` で `Where` を足し、`Count` と行の取得の両方に掛ける。`escapeLike` は `strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")`
- [ ] **Step 4: 通ることを見る** — Run: `cd backend && go test ./internal/infrastructure/postgres/book/ -count=1 -v`（SKIP しないこと）
- [ ] **Step 5: コミット** — `feat: 本の一覧をキーワードと分野タグで絞り込めるようにする`

### Task 3: `ListUsecase` を入力構造体に変える

**Files:**
- Modify: `backend/internal/usecase/book/query/list.go`、`list_test.go`、`get_test.go`（Fake が条件を記録する）

**Interfaces:**
- Consumes: Task 1・2
- Produces: `type ListInput struct { Keyword string; TagID *int64; Limit, Offset int }`、`ListUsecase.Execute(ctx, in ListInput) (*book.BookList, error)`

- [ ] **Step 1: 失敗するテストを書く** — 条件と取得範囲がそのまま渡る。範囲・キーワード・タグIDが不正ならクエリを呼ばない
- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（undefined: query.ListInput）
- [ ] **Step 3: 実装** — `NewListRange` → `NewListCondition` → `FindList`
- [ ] **Step 4: 通ることを見る**
- [ ] **Step 5: コミット** — `feat: 一覧のユースケースに検索・絞り込み条件を渡す`

### Task 4: Handler・OpenAPI・フロント・ドキュメント

**Files:**
- Modify: `backend/internal/presentation/http/book/handler.go`、`handler_test.go`
- Modify: `backend/api/docs/*`・`frontend/web/src/api/generated/*`（生成）
- Modify: `backend/architecture.md`（API 範囲の表）

- [ ] **Step 1: 失敗するテストを書く** — `q`・`tagId` がユースケースに渡る。`q` 101文字・`tagId=0`・`tagId=abc` が400で呼ばない
- [ ] **Step 2: 失敗を確かめる**
- [ ] **Step 3: 実装** — `ListRequest` に `Q string \`query:"q" validate:"max=100"\``・`TagID *int64 \`query:"tagId" validate:"omitempty,min=1"\``。swag の `@Param` に `q`・`tagId`
- [ ] **Step 4: 通ることを見る** — `go build/vet`・`go test ./... -count=1`・`golangci-lint run ./...`
- [ ] **Step 5: コミット** — `feat: 本の一覧APIに検索と分野タグの絞り込みを追加`
- [ ] **Step 6: 再生成してコミット** — `make swagger`、`pnpm gen:api` → `chore: 一覧の検索に合わせてAPIの型を再生成`
- [ ] **Step 7: ドキュメント** — `backend/architecture.md` の `GET /api/books` の行に `q`・`tagId` → `chore: 一覧の検索・絞り込みをドキュメントに反映`
- [ ] **Step 8: ドリフト確認** — `make swagger-check`・`pnpm gen:api:check`・`pnpm typecheck && pnpm lint && pnpm test && pnpm build`
