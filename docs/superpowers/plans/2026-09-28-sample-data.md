# 起動時の見本データ投入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 本が1冊も無い DB で API を起動したとき、実際に読んだ5冊の見本データが自動で入るようにする。

**Architecture:** 見本データ（Go の構造体リテラル）と投入関数 `seedIfEmpty` を `backend/cmd/api/sample_data.go` に置き、`run()` が DB 接続直後に呼ぶ。書誌は見本データ自身が持ち、書影だけ既存の `BookCatalog` に問い合わせる。保存は `book.Repository.Create` に直接書く（登録ユースケースは書誌を外部カタログから取る前提のため使わない）。

**Tech Stack:** Go / Echo / GORM（既存）

**Spec:** `docs/superpowers/specs/2026-09-28-sample-data-design.md`

## Global Constraints

- 見本データに楽天由来の情報（書影 URL・商品ページ URL）を持たせない
- 一言まとめ・感想は「（準備中）」、評価は3（仮の値。本人が書くまで要件定義 §4.3 を満たしていないとデータの冒頭に書く）
- 分野タグは付けない
- 共有のローカル DB を空にしない。テストのダミー ISBN は `978000000330`〜`978000000339` + チェックディジット
- コミットメッセージは日本語・現在形、末尾に `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`

## Review Focus

1. 見本データの1冊でも VO の検証を通らないとき、途中まで保存された状態で起動が止まらない（全冊を先に組み立ててから保存する）。Task 2 で固定する
2. 既に本がある DB で起動しても、外部カタログを1回も呼ばない（初回以外の起動を遅くしない）。Task 2 で固定する
3. 外部カタログの障害で起動が止まらない（書影なしで保存する）。Task 2 で固定する
4. 同時起動で同じ ISBN が先に入っていたら、その冊を飛ばして残りを入れる。Task 2 で固定する
5. 見本データの ISBN が重複していない・書影を持たない。Task 1 で固定する

---

### Task 1: 見本データ

**Files:**
- Create: `backend/cmd/api/sample_data.go`
- Test: `backend/cmd/api/sample_data_test.go`

**Interfaces:**
- Produces: `type sampleBook struct{...}`、`var sampleBooks []sampleBook`、`func (s sampleBook) toBook() (*domainbook.Book, error)`

- [ ] **Step 1: 失敗するテストを書く**（`sample_data_test.go`、`package main`）

```go
func TestSampleBooks_AreValidAndDistinct(t *testing.T) {
	if len(sampleBooks) != 5 {
		t.Fatalf("len(sampleBooks) = %d, want 5", len(sampleBooks))
	}
	seen := map[string]bool{}
	for _, s := range sampleBooks {
		b, err := s.toBook()
		if err != nil {
			t.Fatalf("%s: %v", s.isbn, err)
		}
		if seen[b.ISBN.String()] {
			t.Fatalf("duplicate ISBN %s", b.ISBN.String())
		}
		seen[b.ISBN.String()] = true
		if b.Cover != nil {
			t.Fatalf("%s: sample data must not carry a cover (楽天由来の情報を持たない)", s.isbn)
		}
	}
}
```

- [ ] **Step 2: 失敗を確かめる** — Run: `cd backend && go test ./cmd/api/ -run TestSampleBooks` / Expected: FAIL（`undefined: sampleBooks`）

- [ ] **Step 3: 実装**（`sample_data.go`）

`sampleBook{isbn, title, authors, publisher, publishedOn, titleOverride, summary, comment string; rating int}` と5冊分の `sampleBooks`。書誌は openBD の値（DDIA 第2版は確認できた値だけ）。`toBook` は各 VO で組み立てて `domainbook.New(..., nil, ..., domainbook.TagSelection{})`、`titleOverride` が空でなければ `OverrideTitle`。

- [ ] **Step 4: 通ることを見る** — Expected: PASS

- [ ] **Step 5: コミット** `feat: 起動時に入れる見本データを追加`

### Task 2: `seedIfEmpty`

**Files:**
- Modify: `backend/cmd/api/sample_data.go`、`sample_data_test.go`

**Interfaces:**
- Consumes: Task 1 の `sampleBooks`・`toBook`
- Produces: `func seedIfEmpty(ctx context.Context, books domainbook.Repository, query domainbook.Query, catalog domainbook.BookCatalog) error`

- [ ] **Step 1: 失敗するテストを書く** — 手書き Fake（`fakeSeedBooks`・`fakeSeedQuery`・`fakeSeedCatalog`）で次をテーブル化せずサブテストで確かめる:
  - 空なら全5冊を保存し、カタログの書影を付ける
  - 1冊でもあれば何も保存せず、カタログも呼ばない
  - カタログが障害・該当なしでも書影なしで保存する
  - `Create` が `ErrConflict` の冊は飛ばし、残りを保存して nil を返す
  - `Create` のその他のエラーは返す
  - `FindList` のエラーは返し、何も保存しない

- [ ] **Step 2: 失敗を確かめる** — Expected: FAIL（`undefined: seedIfEmpty`）

- [ ] **Step 3: 実装** — `FindList(ctx, ListRange(1,0))` の `Total` が0でなければ return。全冊を先に `toBook` し、1冊でも失敗ならエラーを返す（何も保存しない）。1冊ずつ `catalog.Lookup` → 成功なら `Cover` を付け、`ErrNotFound` 以外の失敗は warn ログ。`Create` の `ErrConflict` は info ログで飛ばす。

- [ ] **Step 4: 通ることを見る** — Expected: PASS

- [ ] **Step 5: コミット** `feat: 本が無いときだけ見本データを保存する`

### Task 3: 起動時の呼び出しと文書

**Files:**
- Modify: `backend/cmd/api/main.go`、`docs/specifications.md`、`backend/README.md`

- [ ] **Step 1:** `run()` で `bookCatalog := newCatalog(cfg.Catalog)` を作り、DB 接続直後に `seedIfEmpty(context.Background(), pgbook.NewRepository(db), pgbook.NewQuery(db), bookCatalog)` を呼ぶ（エラーなら起動を失敗させる）。`registerRoutes` には同じ `bookCatalog` を渡す
- [ ] **Step 2:** `go build ./... && go test ./cmd/api/ -count=1` — Expected: PASS（既存の `TestRun_*` も通る）
- [ ] **Step 3:** 要件定義 §4.3 の「4冊」→「5冊」、冊の一覧を spec §1 の5冊に。README の起動手順に「本が1冊も無ければ起動時に見本データ（5冊）が入る」を1行
- [ ] **Step 4: コミット** `feat: 起動時に見本データを投入する` / `chore: 見本データを5冊にし起動手順に追記`
- [ ] **Step 5:** `go vet ./...`・`go test ./... -count=1`・`golangci-lint run ./...`・`make swagger-check`
