# 書名の上書き・一言まとめ 実装計画

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 読んだ本に「書名の上書き」と「一言まとめ」を持たせ、登録・更新の API で受け取り、詳細・一覧で返す。

**Architecture:** Domain に `Title`・`Summary` の VO を足し、`Book` 集約に `TitleOverride`・`Summary` を持たせる。永続化はマイグレーション 000002 で `book` に列を足し、Repository/Query・ユースケース・Handler を縦に通す。表示する書名は Query で決め、Read Model には文字列だけを載せる。

**Tech Stack:** Go 1.x / Echo / GORM / golang-migrate / swag / Orval（フロントの型の再生成のみ）

**Spec:** `docs/superpowers/specs/2026-09-28-title-override-summary-design.md`（変更前の形は `docs/superpowers/specs/2026-09-28-reading-api-design.md`）

## Global Constraints

- 書名の上書き: 1〜255 文字、前後の空白を除く、タブ・改行などの制御文字は不可（トリム前に検査）。前後の空白を除いて空なら「上書きなし」（エラーにしない）
- 一言まとめ: 1〜100 文字、前後の空白を除く、改行を含む制御文字は不可、必須
- 外部カタログを取り直しても書名の上書きは消さない
- PUT は全体の置き換え。`titleOverride` を省略・空にすると上書きを外す
- 入力の検証は外部カタログへの問い合わせより先に済ませる
- Read Model は VO 型を返さない（`string` / `*string`）
- コメントは日本語。Repository の型と各メソッドには1行の What コメント（CLAUDE.md）
- 関数の循環的複雑度は 10 以下（gocyclo / ESLint。テストは対象外）
- テスト名は期待する振る舞いを日本語で書く。境界値は「上限ちょうど」と「上限+1」の両方
- コミットは1つの論理的な変更ごとに分け、各コミット単体で `go build ./... && go test ./...` が通ること

## Review Focus

1. 更新で `titleOverride` を送らない → 既存の上書きが外れる（PUT の置き換え）。サイクル 5 のテストで固定する
2. 更新で外部カタログが別の書名を返す → 上書きがあれば表示する書名は上書きのまま。サイクル 1・5 のテストで固定する
3. 一言まとめに改行を含める → 400（感想は改行可なので取り違えやすい）。Task 2 のテストで固定する
4. 書名の上書きに空白だけを送る → エラーにせず上書きなし。サイクル 4 のテストで固定する

## PR の分け方（設計書 §8 からの変更）

設計書 §8 は「ドメイン」と「それ以外」の2本としていた。ただし `Book.New`・`ChangeReview` の引数を変えると、呼び出し側（Repository・ユースケース）も同じコミットで直さないとビルドが通らない。そこで次の形にする。

| PR | タスク | 中身 |
|---|---|---|
| 1（ドメイン） | Task 1・2 | 呼び出し側を壊さない VO（`Title`・`Summary`）だけ。設計書と計画も含める |
| 2（それ以外） | Task 3〜5 | マイグレーション、`Book` 集約の変更と各層の配線、OpenAPI・フロントの型・ドキュメント |

---

### Task 1: `Title` VO を足す（PR 1）

**Files:**
- Create: `backend/internal/domain/book/title.go`
- Create: `backend/internal/domain/book/title_test.go`

**Interfaces:**
- Produces: `func NewTitle(raw string) (Title, error)`、`func (t Title) String() string`（検証は `bibliography.go` の `text` と `titleMaxLength` を使う。`Bibliography` は変えない）

- [ ] **Step 1: 失敗するテストを書く**（`title_test.go`）

```go
package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "前後の空白を除いて保持する", in: " 徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版 ", want: "徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版"},
		{name: "途中の全角空白は残す", in: "改訂新版　良いコード／悪いコードで学ぶ設計入門", want: "改訂新版　良いコード／悪いコードで学ぶ設計入門"},
		{name: "1文字は有効", in: "本", want: "本"},
		{name: "255文字ちょうどは有効", in: strings.Repeat("あ", 255), want: strings.Repeat("あ", 255)},
		{name: "256文字はエラー", in: strings.Repeat("あ", 256), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "　 ", wantErr: true},
		{name: "改行はエラー", in: "人間\n失格", wantErr: true},
		{name: "前後のタブもエラー", in: "\t人間失格", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewTitle(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 失敗を確かめる**

Run: `cd backend && go test ./internal/domain/book/ -run TestNewTitle`
Expected: FAIL（`undefined: book.NewTitle`）

- [ ] **Step 3: 実装する**（`title.go`）

```go
package book

// Title は自分で上書きする書名の VO。規則は外部カタログの書名（Bibliography）と同じなので、同じ検証を使う。
type Title struct {
	value string
}

// NewTitle は前後の空白を除いて1〜255文字の書名を作る。タブや改行などの制御文字は受け付けない。
func NewTitle(raw string) (Title, error) {
	v, err := text("書名", raw, 1, titleMaxLength)
	if err != nil {
		return Title{}, err
	}
	return Title{value: v}, nil
}

func (t Title) String() string {
	return t.value
}
```

- [ ] **Step 4: 通ることを確かめる**

Run: `cd backend && go test ./internal/domain/book/ && go build ./...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add backend/internal/domain/book/title.go backend/internal/domain/book/title_test.go
git commit -m "feat: 書名の上書きに使う書名のVOを追加"
```

### Task 2: `Summary` VO を足す（PR 1）

**Files:**
- Create: `backend/internal/domain/book/summary.go`
- Create: `backend/internal/domain/book/summary_test.go`

**Interfaces:**
- Produces: `func NewSummary(raw string) (Summary, error)`、`func (s Summary) String() string`

- [ ] **Step 1: 失敗するテストを書く**（`summary_test.go`）

```go
package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewSummary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "前後の空白を除いて保持する", in: "  分散データの設計を体系的に学べる  ", want: "分散データの設計を体系的に学べる"},
		{name: "1文字は有効", in: "良", want: "良"},
		{name: "100文字ちょうどは有効", in: strings.Repeat("あ", 100), want: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", in: strings.Repeat("あ", 101), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "  ", wantErr: true},
		{name: "改行はエラー（感想と違い1行の文）", in: "設計\n入門", wantErr: true},
		{name: "前後の改行もエラー", in: "設計入門\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewSummary(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 失敗を確かめる**

Run: `cd backend && go test ./internal/domain/book/ -run TestNewSummary`
Expected: FAIL（`undefined: book.NewSummary`）

- [ ] **Step 3: 実装する**（`summary.go`）

```go
package book

const summaryMaxLength = 100

// Summary は一覧で見せる一言まとめの VO。1行の短い文なので、感想と違い改行も受け付けない。
type Summary struct {
	value string
}

// NewSummary は前後の空白を除いて1〜100文字の一言まとめを作る。
func NewSummary(raw string) (Summary, error) {
	v, err := text("一言まとめ", raw, 1, summaryMaxLength)
	if err != nil {
		return Summary{}, err
	}
	return Summary{value: v}, nil
}

func (s Summary) String() string {
	return s.value
}
```

- [ ] **Step 4: 通ることを確かめる**

Run: `cd backend && go test ./internal/domain/book/ && go vet ./...`
Expected: PASS

- [ ] **Step 5: コミット**

```bash
git add backend/internal/domain/book/summary.go backend/internal/domain/book/summary_test.go
git commit -m "feat: 一言まとめのVOを追加"
```

PR 1 を作る（ベース `develop`、設計書・計画・Task 1・2 のコミット）。

---

### Task 3: マイグレーション 000002 と DB の資料（PR 2）

**Files:**
- Create: `backend/migrations/000002_add_book_title_override_and_summary.up.sql`
- Create: `backend/migrations/000002_add_book_title_override_and_summary.down.sql`
- Modify: `docs/db/backend-schema.json`、`docs/db/backend-schema.md`

- [ ] **Step 1: up を書く**

```sql
ALTER TABLE book ADD COLUMN title_override VARCHAR(255);
-- 既存の行があっても足せるよう既定値付きで足し、すぐ外す（以後は必ずアプリが値を入れる）
ALTER TABLE book ADD COLUMN summary VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE book ALTER COLUMN summary DROP DEFAULT;

COMMENT ON COLUMN book.title_override IS '自分で上書きした書名（1〜255文字）。NULLなら上書きしていない。外部カタログを取り直しても消えない';
COMMENT ON COLUMN book.summary IS '一覧で見せる一言まとめ（1〜100文字、改行なし）';
COMMENT ON COLUMN book.title IS '書名（外部カタログの値）。画面では title_override があればそちらを見せる';
```

- [ ] **Step 2: down を書く**

```sql
COMMENT ON COLUMN book.title IS '書名（外部カタログの値）';
ALTER TABLE book DROP COLUMN IF EXISTS summary;
ALTER TABLE book DROP COLUMN IF EXISTS title_override;
```

- [ ] **Step 3: ローカル DB で up / down / up を確かめる**

Run: `cd backend && make migrate-up && make migrate-down && make migrate-up`（`migrate-down` は1つだけ戻す）
Expected: いずれもエラーなし。`psql` の `\d+ book` に2列とコメントがある

- [ ] **Step 4: DB の資料を更新する**

`docs/db/backend-schema.json` の `book.columns` の `title` の次に足し、`title` の説明を揃え、`source` に 000002 を足す。

```json
{ "name": "title_override", "type": "VARCHAR(255)", "nullable": true, "description": "自分で上書きした書名（1〜255文字）。NULLなら上書きしていない。外部カタログを取り直しても消えない" },
```

`rating` の次に足す。

```json
{ "name": "summary", "type": "VARCHAR(100)", "nullable": false, "description": "一覧で見せる一言まとめ（1〜100文字、改行なし）" },
```

`docs/db/backend-schema.md` の ER 図の `book` に `varchar_255 title_override "nullable"` と `varchar_100 summary` を足す。

- [ ] **Step 5: コミット**

```bash
git add backend/migrations/000002_* docs/db/backend-schema.json docs/db/backend-schema.md
git commit -m "feat: 読んだ本に書名の上書きと一言まとめの列を追加"
```

### Task 4: 一言まとめと書名の上書きを各層に通す（PR 2）

`Book.New`・`ChangeReview` の引数が変わるため、Domain・Repository・Query・ユースケース・Handler を同じコミットで直す（どれか1層だけではビルドが通らない）。

進め方は TDD（superpowers:test-driven-development）のサイクルを、振る舞いごとに下の層から回す。各サイクルは「失敗するテストを書く → **失敗を実際に見る** → 最小の実装 → 通ることを見る」。コミットはビルドが通る最後の1回だけにする（サイクルは細かく、コミットはまとめる）。

| サイクル | 振る舞い |
|---|---|
| 1 | Domain: 一言まとめを持ち、書名を上書き・外せ、取り直しても上書きが残る |
| 足場 | 呼び出し側をビルドが通る形にする（新しい振る舞いは足さない） |
| 2 | Repository: 一言まとめと上書きを保存して読める。上書きを外すと NULL |
| 3 | Query: 表示する書名は上書き優先。詳細は上書きも返す |
| 4 | 登録: 一言まとめと上書きを付けて登録し、不正ならカタログも保存も呼ばない |
| 5 | 更新: 一言まとめと上書きを差し替え、空なら上書きを外し、取り直しても上書きは残る |
| 6 | Handler: 入力を Command に渡し、欠落・上限超えを 400 にし、応答に新しい項目を載せる |

**Files:**
- Modify: `backend/internal/domain/book/book.go`、`book_test.go`
- Modify: `backend/internal/infrastructure/postgres/book/repository.go`、`query.go`、`repository_test.go`、`query_test.go`
- Modify: `backend/internal/usecase/book/command/register.go`、`update.go`、`register_test.go`、`update_test.go`
- Modify: `backend/internal/presentation/http/book/handler.go`、`handler_test.go`

**Interfaces:**
- Consumes: `NewTitle`、`NewSummary`（Task 1・2）
- Produces:
  - `func New(isbn ISBN, bibliography Bibliography, cover *Cover, summary Summary, comment Comment, rating Rating) *Book`
  - `func (b *Book) ChangeReview(summary Summary, comment Comment, rating Rating, version int)`
  - `func (b *Book) OverrideTitle(title *Title)`
  - `Book.TitleOverride *Title`、`Book.Summary Summary`
  - `BookDetail` に `TitleOverride *string`・`Summary string`、`BookListItem` に `Summary string`（`Title` はどちらも表示する書名）
  - `RegisterCommand`・`UpdateCommand` に `Summary string`・`TitleOverride string`
  - `RegisterRequest`・`UpdateRequest` に `Summary string`（`required,max=100`）・`TitleOverride string`（`max=255`）
  - `Response` に `Summary`・`TitleOverride *string`、`ListItemResponse` に `Summary`

#### サイクル 1: Domain

- [ ] **Step 1: 失敗するテストを書く**（`book_test.go`）

`mustBook` に `summary, _ := book.NewSummary("分散データの設計を体系的に学べる")` を足し、`book.New(isbn, bib, &cover, summary, comment, rating)` にする。次のテストを足す・直す。

```go
func TestBook_ChangeReview(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	summary, _ := book.NewSummary("読み返して見方が変わった")
	comment, _ := book.NewComment("読み返して評価が変わった")
	rating, _ := book.NewRating(4)

	b.ChangeReview(summary, comment, rating, 3)

	if b.Summary != summary || b.Comment != comment || b.Rating != rating || b.Version != 3 {
		t.Fatalf("got summary=%q comment=%q rating=%d version=%d", b.Summary.String(), b.Comment.String(), b.Rating.Int(), b.Version)
	}
	if b.Bibliography.Title() != "データ指向アプリケーションデザイン" || b.Cover == nil {
		t.Fatal("changing the review must not touch the catalog data")
	}
}

func TestBook_OverrideTitle(t *testing.T) {
	t.Parallel()

	t.Run("上書きを付けるとコピーして持つ", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := book.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		title = book.Title{}
		if b.TitleOverride == nil || b.TitleOverride.String() != "正しい書名" {
			t.Fatalf("TitleOverride = %v, want a copy of the given title", b.TitleOverride)
		}
	})

	t.Run("nilで上書きを外す", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := book.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		b.OverrideTitle(nil)
		if b.TitleOverride != nil {
			t.Fatalf("TitleOverride = %v, want nil", b.TitleOverride)
		}
	})

	t.Run("外部カタログを取り直しても上書きは残る", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := book.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		bib, _ := book.NewBibliography("カタログの別の書名", "", "", "")
		b.RefreshCatalog(bib, nil)
		if b.TitleOverride == nil || b.TitleOverride.String() != "正しい書名" {
			t.Fatalf("TitleOverride = %v, want it kept after refreshing the catalog", b.TitleOverride)
		}
	})
}
```

`TestNew_SetsFields...` に `b.Summary.String() != "分散データの設計を体系的に学べる"` と `b.TitleOverride != nil` の検査を足す（新規作成時は上書きなし）。

- [ ] **Step 2: 失敗を見る**

Run: `cd backend && go test ./internal/domain/book/`
Expected: FAIL（`too many arguments in call to book.New`、`b.OverrideTitle undefined` など）

- [ ] **Step 3: 最小の実装**（`book.go`）

```go
type Book struct {
	ID           ID
	ISBN         ISBN
	Bibliography Bibliography
	// Cover は書影。提供元に書影が無い本は nil。
	Cover *Cover
	// TitleOverride は自分で上書きした書名。上書きしていなければ nil。外部カタログを取り直しても消さない。
	TitleOverride *Title
	Summary       Summary
	Comment       Comment
	Rating        Rating
	Version       int
}

func New(isbn ISBN, bibliography Bibliography, cover *Cover, summary Summary, comment Comment, rating Rating) *Book {
	return &Book{ISBN: isbn, Bibliography: bibliography, Cover: copyOf(cover), Summary: summary, Comment: comment, Rating: rating}
}

// ChangeReview は一言まとめ・感想・評価を差し替える。version は更新元が読んだバージョン（楽観的ロックに使う）。
func (b *Book) ChangeReview(summary Summary, comment Comment, rating Rating, version int) {
	b.Summary = summary
	b.Comment = comment
	b.Rating = rating
	b.Version = version
}

// OverrideTitle は書名を自分で上書きする。nil なら上書きを外し、外部カタログの書名に戻す。
func (b *Book) OverrideTitle(title *Title) {
	b.TitleOverride = copyOf(title)
}
```

Read Model を変える。

```go
type BookDetail struct {
	ID   ID
	ISBN string
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title string
	// TitleOverride は自分で上書きした書名。上書きしていなければ nil（編集画面が今の上書きを送り直すのに使う）。
	TitleOverride *string
	Summary       string
	// （Authors 以降は今のまま）
}

type BookListItem struct {
	ID   ID
	ISBN string
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title   string
	Summary string
	// （Authors 以降は今のまま）
}
```

- [ ] **Step 4: 通ることを見る**

Run: `cd backend && go test ./internal/domain/book/`
Expected: PASS（この時点で他のパッケージはビルドが通らない。次の足場で直す）

#### 足場: 呼び出し側をビルドが通る形にする

新しい振る舞いは足さない。この後のサイクルで「失敗を見る」ために、値の受け渡しはまだ繋がない。

- [ ] **Step 5: 型と引数だけを合わせる**

- `repository.go` の `model` に `TitleOverride *string \`gorm:"column:title_override;size:255"\`` と `Summary string \`gorm:"column:summary;size:100;not null"\`` を足す。`adapt` は `summary, err := domainbook.NewSummary(row.Summary)` を作って `domainbook.New(isbn, bib, cover, summary, comment, rating)` に渡す（`toModel` と `Update` にはまだ足さない）
- `register.go` は `book.New(isbn, entry.Bibliography, entry.Cover, book.Summary{}, comment, rating)`、`update.go` は `b.ChangeReview(b.Summary, comment, rating, cmd.Version)` にする（サイクル 4・5 で置き換える仮の値）
- テストの呼び出し（`repository_test.go` の `newBook`、`update_test.go` の `existingBook`）に `NewSummary` の値を渡し、`ChangeReview` に第1引数を足す

- [ ] **Step 6: 既存のテストが通ることを見る**

Run: `cd backend && go build ./... && go test ./...`
Expected: PASS（新しいテストはまだ無い）

#### サイクル 2: Repository

- [ ] **Step 7: 失敗するテストを書く**（`repository_test.go`）


足場で `newBook` に渡した一言まとめは `"一言まとめ"` にしておく。次を足す。

```go
func TestRepository_TitleOverrideAndSummary(t *testing.T) {
	db := connectTestDB(t)
	repo := pgbook.NewRepository(db)

	t.Run("書名の上書きと一言まとめを保存して読み込める", func(t *testing.T) {
		b := newBook(t, "9780000000101", "カタログの書名", nil, 4)
		title, _ := domainbook.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		createBook(t, db, b)

		got, err := repo.FindByID(context.Background(), b.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.TitleOverride == nil || got.TitleOverride.String() != "正しい書名" || got.Summary.String() != "一言まとめ" {
			t.Fatalf("got override=%v summary=%q", got.TitleOverride, got.Summary.String())
		}
	})

	t.Run("上書きを外して更新するとNULLになる", func(t *testing.T) {
		b := newBook(t, "9780000000102", "カタログの書名", nil, 4)
		title, _ := domainbook.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		createBook(t, db, b)

		b.OverrideTitle(nil)
		if err := repo.Update(context.Background(), b); err != nil {
			t.Fatalf("Update: %v", err)
		}
		var override *string
		db.Raw("SELECT title_override FROM book WHERE id = ?", int64(b.ID)).Scan(&override)
		if override != nil {
			t.Fatalf("title_override = %q, want NULL", *override)
		}
	})
}
```

- [ ] **Step 8: 失敗を見る**

Run: `cd backend && make migrate-up && go test ./internal/infrastructure/postgres/book/ -run TestRepository_TitleOverrideAndSummary -v`
Expected: 「保存して読み込める」が FAIL（`toModel` が一言まとめを書かないので空文字が入り、読み込みが ErrInvalid になる）。「上書きを外すと NULL」はこの時点で PASS してよい（上書きをまだ書かないため。実装後に Update が NULL を書くことを守るテスト）。SKIP になっていないこと

- [ ] **Step 9: 最小の実装**

`Update` の `Updates(map[string]any{...})` に `"title_override": row.TitleOverride, "summary": row.Summary` を足す。`toModel` に足す。

```go
		Summary:     b.Summary.String(),
	}
	if b.TitleOverride != nil {
		v := b.TitleOverride.String()
		row.TitleOverride = &v
	}
```

`adapt` の複雑度が 10 を超えないよう、書名の上書きの復元を関数に分ける。

```go
// adaptTitleOverride は保存済みの上書きを VO で検証し直す。NULL なら上書きなし。
func adaptTitleOverride(v *string) (*domainbook.Title, error) {
	if v == nil {
		return nil, nil
	}
	t, err := domainbook.NewTitle(*v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
```

`adapt` に `override, err := adaptTitleOverride(row.TitleOverride)` を足し（一言まとめは足場で済んでいる）、`domainbook.New(...)` の後に `b.OverrideTitle(override)` を呼ぶ。

- [ ] **Step 10: 通ることを見る**

Run: `cd backend && go test ./internal/infrastructure/postgres/book/ -run TestRepository -v`
Expected: PASS

#### サイクル 3: Query

- [ ] **Step 11: 失敗するテストを書く**（`query_test.go`）

`query_test.go` に次を足す（既存テストの ISBN と重ならない値を使う）。

```go
func TestQuery_DisplayTitle(t *testing.T) {
	db := connectTestDB(t)
	q := pgbook.NewQuery(db)

	overridden := newBook(t, "9780000000201", "カタログの書名", nil, 4)
	title, _ := domainbook.NewTitle("正しい書名")
	overridden.OverrideTitle(&title)
	createBook(t, db, overridden)
	plain := newBook(t, "9780000000202", "上書きなしの書名", nil, 4)
	createBook(t, db, plain)

	t.Run("詳細は上書きを優先し、上書きも返す", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), overridden.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.Title != "正しい書名" || got.TitleOverride == nil || *got.TitleOverride != "正しい書名" || got.Summary != "一言まとめ" {
			t.Fatalf("got title=%q override=%v summary=%q", got.Title, got.TitleOverride, got.Summary)
		}
	})

	t.Run("上書きが無ければカタログの書名を表示し、上書きはnil", func(t *testing.T) {
		got, err := q.FindDetailByID(context.Background(), plain.ID)
		if err != nil {
			t.Fatalf("FindDetailByID: %v", err)
		}
		if got.Title != "上書きなしの書名" || got.TitleOverride != nil {
			t.Fatalf("got title=%q override=%v", got.Title, got.TitleOverride)
		}
	})

	t.Run("一覧も表示する書名と一言まとめを返す", func(t *testing.T) {
		r, _ := domaincommon.NewListRange(100, 0)
		list, err := q.FindList(context.Background(), r)
		if err != nil {
			t.Fatalf("FindList: %v", err)
		}
		titles := map[domainbook.ID]*domainbook.BookListItem{}
		for _, it := range list.Items {
			titles[it.ID] = it
		}
		if it := titles[overridden.ID]; it == nil || it.Title != "正しい書名" || it.Summary != "一言まとめ" {
			t.Fatalf("overridden item = %+v", it)
		}
		if it := titles[plain.ID]; it == nil || it.Title != "上書きなしの書名" {
			t.Fatalf("plain item = %+v", it)
		}
	})
}
```

既存の Query テストの Read Model の型検査（`Title` が `string` であること）に `Summary`・`TitleOverride` の型も足す。

- [ ] **Step 12: 失敗を見る**

Run: `cd backend && go test ./internal/infrastructure/postgres/book/ -run TestQuery_DisplayTitle -v`
Expected: FAIL（`Title` がカタログの書名のまま、`Summary` が空）

- [ ] **Step 13: 最小の実装**

`query.go` に表示する書名の決め方を1か所に置き、詳細と一覧の両方で使う。

```go
// displayTitle は画面に出す書名を決める。自分で上書きした書名があればそれ、無ければ外部カタログの書名。
func displayTitle(override *string, catalogTitle string) string {
	if override != nil {
		return *override
	}
	return catalogTitle
}
```

詳細は `Title: displayTitle(row.TitleOverride, row.Title), TitleOverride: row.TitleOverride, Summary: row.Summary` を足す。一覧は `Select` に `title_override, summary` を足し、`Title: displayTitle(row.TitleOverride, row.Title), Summary: row.Summary` にする。


- [ ] **Step 14: 通ることを見る**

Run: `cd backend && go test ./internal/infrastructure/postgres/book/ -v`
Expected: PASS

#### サイクル 4: 登録のユースケース

- [ ] **Step 15: 失敗するテストを書く**（`register_test.go`）

- `valid := command.RegisterCommand{ISBN: "4873118700", Summary: "分散データの設計を学べる", Comment: "良書", Rating: 5}`
- 正常系の検査に `c.Summary.String() != "分散データの設計を学べる"` と `c.TitleOverride != nil` を足す
- `invalid` に足す: `{name: "空の一言まとめ", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: " ", Comment: "良書", Rating: 5}}`、`{name: "改行を含む一言まとめ", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "a\nb", Comment: "良書", Rating: 5}}`、`{name: "256文字の書名の上書き", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "要約", TitleOverride: strings.Repeat("あ", 256), Comment: "良書", Rating: 5}}`。既存の3件にも `Summary: "要約"` を足す
- 追加: 「書名の上書きを指定すると付けて登録する」（`TitleOverride: "正しい書名"` → `books.created.TitleOverride.String() == "正しい書名"`）
- 追加: 「空白だけの書名の上書きはエラーにせず上書きなしで登録する」（`TitleOverride: "　"` → エラーなし、`books.created.TitleOverride == nil`）

- [ ] **Step 16: 失敗を見る**

Run: `cd backend && go test ./internal/usecase/book/command/ -run TestRegisterUsecase -v`
Expected: FAIL（`unknown field Summary in struct literal` の後、フィールドを足すと一言まとめが空・上書きが付かない・不正な入力でカタログが呼ばれる）

- [ ] **Step 17: 最小の実装**

両方の Command に足す。

```go
	Summary string
	// TitleOverride は自分で上書きする書名。前後の空白を除いて空なら上書きしない（更新では上書きを外す）。
	TitleOverride string
```

`command` パッケージに入力の解釈を1か所に置く（登録と更新で同じ規則のため）。新しいファイル `backend/internal/usecase/book/command/title_override.go`:

```go
package command

import (
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// parseTitleOverride は書名の上書きの入力を解釈する。前後の空白を除いて空なら「上書きしない」、それ以外は書名の規則で検証する。
func parseTitleOverride(raw string) (*book.Title, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := book.NewTitle(raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
```

`register.go`: ISBN の検証の後に `summary, err := book.NewSummary(cmd.Summary)` と `override, err := parseTitleOverride(cmd.TitleOverride)` を足し（どちらもカタログの前）、`b := book.New(isbn, entry.Bibliography, entry.Cover, summary, comment, rating)` の後に `b.OverrideTitle(override)`。複雑度が 10 を超える場合は、入力の検証を `validateRegister(cmd RegisterCommand) (registerInput, error)` のような非公開関数に分ける。

- [ ] **Step 18: 通ることを見る**

Run: `cd backend && go test ./internal/usecase/book/command/ -run TestRegisterUsecase -v`
Expected: PASS

#### サイクル 5: 更新のユースケース

- [ ] **Step 19: 失敗するテストを書く**（`update_test.go`）

- `existingBook` で `summary, _ := book.NewSummary("最初のまとめ")` を使い、`title, _ := book.NewTitle("前の上書き"); b.OverrideTitle(&title)` で上書き付きにする
- `cmd := command.UpdateCommand{ID: 10, Summary: "読み返したまとめ", TitleOverride: "新しい上書き", Comment: "読み返した", Rating: 5, Version: 2}`
- 正常系の検査に `u.Summary.String() != "読み返したまとめ"`、`u.TitleOverride.String() != "新しい上書き"` を足し、カタログが「新しい書名」を返しても `u.TitleOverride` は「新しい上書き」のままであることを確かめる
- 追加: 「書名の上書きを空で送ると上書きを外す」（`TitleOverride: ""` → `books.updated.TitleOverride == nil`）
- 追加: 「空の一言まとめはカタログも保存も呼ばずにエラー」（`Summary: ""` → `ErrInvalid`、`catalog.called == 0`、`books.updated == nil`）

- [ ] **Step 20: 失敗を見る**

Run: `cd backend && go test ./internal/usecase/book/command/ -run TestUpdateUsecase -v`
Expected: FAIL（一言まとめが差し替わらない、上書きが変わらない、空の一言まとめでカタログが呼ばれる）

- [ ] **Step 21: 最小の実装**

`update.go`: `FindByID` の後に同様に検証し、`b.ChangeReview(summary, comment, rating, cmd.Version)` と `b.OverrideTitle(override)` を呼ぶ（`refreshCatalog` の前後どちらでもよいが、上書きには触れないこと）。

- [ ] **Step 22: 通ることを見る**

Run: `cd backend && go test ./internal/usecase/book/command/ -v`
Expected: PASS

#### サイクル 6: Handler

- [ ] **Step 23: 失敗するテストを書く**

`detail`・`detailResponse` に `Summary: "分散データの設計を学べる"` を足す。一覧の項目にも `Summary` を足す。登録・更新のリクエストの JSON に `"summary":"分散データの設計を学べる"` を足し、Command に `Summary` と `TitleOverride` が渡ることを確かめる。次のケースを足す（既存の 400 のテーブルに入れる）。

| ケース | リクエスト | 期待 |
|---|---|---|
| 一言まとめが無い | `{"isbn":"9784873118703","comment":"良書","rating":5}` | 400、`errors` に `{"field":"summary","rule":"required"}` |
| 一言まとめが101文字 | `summary` に101文字 | 400、`{"field":"summary","rule":"max"}` |
| 書名の上書きが256文字 | `titleOverride` に256文字 | 400、`{"field":"titleOverride","rule":"max"}` |
| 書名の上書きを省略 | `titleOverride` なし | 200、Command の `TitleOverride` が `""` |

- [ ] **Step 24: 失敗を見る**

Run: `cd backend && go test ./internal/presentation/http/book/ -v`
Expected: FAIL（`summary` が無くても 400 にならない、Command に `Summary` が渡らない、応答に `summary` が無い）

- [ ] **Step 25: 最小の実装**

```go
type RegisterRequest struct {
	ISBN    string `json:"isbn" validate:"required,max=17" example:"9784873118703"`
	Summary string `json:"summary" validate:"required,max=100" example:"分散データの設計を体系的に学べる"`
	// TitleOverride は外部カタログの書名が実際と違うときに自分で付ける書名。省略・空（空白だけも）なら上書きしない。
	TitleOverride string `json:"titleOverride" validate:"max=255" example:""`
	Comment       string `json:"comment" validate:"required,max=5000" example:"分散システムの設計を体系的に学べた"`
	Rating        *int   `json:"rating" validate:"required,min=1,max=5" example:"5"`
} // @name RegisterBookRequest

type UpdateRequest struct {
	Summary string `json:"summary" validate:"required,max=100" example:"読み返して理解が深まった"`
	// TitleOverride は全体の置き換えなので、省略・空（空白だけも）なら上書きを外す。
	TitleOverride string `json:"titleOverride" validate:"max=255" example:""`
	Comment       string `json:"comment" validate:"required,max=5000" example:"読み返して理解が深まった"`
	Rating        *int   `json:"rating" validate:"required,min=1,max=5" example:"5"`
	Version       *int   `json:"version" validate:"required" example:"1"`
} // @name UpdateBookRequest
```

`Response` に足す（`Title` のコメントも直す）。

```go
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title string `json:"title" example:"データ指向アプリケーションデザイン"`
	// TitleOverride は自分で上書きした書名。上書きしていなければ null。
	TitleOverride *string `json:"titleOverride" example:"徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版"`
	Summary       string  `json:"summary" example:"分散データの設計を体系的に学べる"`
```

`ListItemResponse` に `Summary string \`json:"summary" example:"分散データの設計を体系的に学べる"\`` を足す。`RegisterBook`・`Update` で Command に `Summary: req.Summary, TitleOverride: req.TitleOverride` を渡し、`toResponse`・一覧の変換に新しい項目を足す。swag の `@Description` に「`titleOverride` を省略・空にすると上書きを外す」（更新）を足す。

- [ ] **Step 26: 通ることを見る**

Run: `cd backend && go test ./internal/presentation/http/book/ -v`
Expected: PASS

#### 仕上げ

- [ ] **Step 27: すべて通ることを見る**

Run: `cd backend && make migrate-up && go build ./... && go test ./... && golangci-lint run ./...`
Expected: PASS、`0 issues.`（契約テストが skip されていないこと: `go test -v ./internal/infrastructure/postgres/... | grep -c SKIP` が 0）

- [ ] **Step 28: コミット**

```bash
git add backend/internal
git commit -m "feat: 書名の上書きと一言まとめを登録・更新で受け取り、詳細・一覧で返す"
```

### Task 5: OpenAPI・フロントの型・ドキュメント（PR 2）

**Files:**
- Modify: `backend/api/docs/swagger.json`、`swagger.yaml`（生成）
- Modify: `frontend/web/src/api/generated/*`（生成）
- Modify: `backend/architecture.md`（プロダクト API 範囲の表）、`backend/README.md`（curl の例）

- [ ] **Step 1: 再生成する**

Run: `cd backend && make swagger && cd ../frontend && pnpm gen:api`
Expected: `swagger.yaml` に `summary`・`titleOverride` が出る

- [ ] **Step 2: フロントの検査を通す**

Run: `cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build`
Expected: PASS

- [ ] **Step 3: ドキュメントを直す**

- `backend/architecture.md` の `POST /api/books` の用途を「読んだ本を登録する（`isbn`・`summary`・`comment`・`rating`、任意で `titleOverride`。…）」に、`PUT /api/books/{id}` を「一言まとめ・感想・評価・書名の上書きを更新する（楽観的ロック。書誌と書影を取り直す）」にする
- `backend/README.md` の curl の例の JSON に `"summary":"要約"` を足す

- [ ] **Step 4: ずれが無いことを確かめる**

Run: `cd backend && make swagger-check && cd ../frontend && pnpm gen:api:check`
Expected: 差分なし

- [ ] **Step 5: コミット（2つに分ける）**

```bash
git add backend/api/docs frontend/web/src/api/generated
git commit -m "chore: 書名の上書きと一言まとめに合わせてAPIの型を再生成"
git add backend/architecture.md backend/README.md
git commit -m "chore: 書名の上書きと一言まとめをドキュメントに反映"
```

PR 2 を作る（ベースは PR 1 のブランチ、PR 1 のマージ後に `develop` へ切り替える）。
