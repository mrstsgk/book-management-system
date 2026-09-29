# 公開画面（単位2） Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement each PR's task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 一覧・詳細・分野別の集計の3画面を `frontend/web` に作る。

**Architecture:** 共通の土台（ルーティング・レイアウト・画面の状態の部品・評価・一覧の infinite query 生成）を PR1 で先に入れ、マージ後に3画面を feature ごとに別 PR で並行して作る（`features/books-list`・`features/book-detail`・`features/tag-counts`。feature 間 import はしない）。

**Tech Stack:** React 19 / React Router 7 / TanStack Query 5（Orval 生成）/ Tailwind 3 / Vitest + RTL + MSW

**Spec:** `docs/superpowers/specs/2026-09-29-public-screens-design.md`（見た目は `docs/design/public-screens/*.html`）

## Global Constraints

- 生成物（`web/src/api/generated/`）は手編集しない。`pnpm gen:api` で作り直す
- Query フックを手書きしない（Orval 生成のフックを feature の hook から呼ぶ）
- 配色は既存の `brand`・`ink`。主色の文字・ボタンは `brand-700`（`#bf3f1f`。白地で 4.5:1 を満たす）
- テスト名は日本語で期待する振る舞いを書く。操作は `userEvent`、クエリは `getByRole` 等を優先
- PR ごとに `cd frontend && pnpm typecheck && pnpm lint && pnpm format:check && pnpm test && pnpm build && pnpm build-storybook` と `pnpm gen:api:check` を通す
- ブランチ・コミット・PR は `docs/rules/git-*.md` に従う。push は `gh auth status` の有効アカウントが `mrstsgk` のときだけ行い、違えば止めて報告する（資格情報を取り出して回避しない）

## Review Focus

- `tag=abc` や `tag=0` の URL を直接開く → 絞り込みなしで一覧が出る（400 を出さない）
- 検索中に「もっと見る」で2ページ目を読んだあと、キーワードを変える → 先頭20冊から出し直し、古い2ページ目が混ざらない
- 楽天の書影なのに `coverProductUrl` が空 → 商品ページのリンクを壊れた href で出さない（リンクだけ省き、クレジットと注記は出す）
- 感想の改行・長い書名 → 改行が保たれ、長い書名で横スクロールが出ない（スマホ幅）
- 詳細の id が数値でない（`/books/abc`）→ API を呼ばずに見つからない表示

---

### PR1: 土台（`feat/public-screens-foundation`）

**Files:**
- Modify: `frontend/web/orval.config.ts`（`override.operations.getApiBooks.query = { useInfinite: true, useInfiniteQueryParam: 'offset' }`）→ `pnpm gen:api`
- Modify: `web/src/app/router.tsx`（`/`・`/books/:id`・`/tags`・`*`。3画面は PR2〜4 まで「準備中」の仮ページ `web/src/app/routes/*.tsx` を置き、各 PR がその1ファイルだけを差し替える）
- Modify: `web/src/components/layouts/Header.tsx`・`Footer.tsx`
- Create: `web/src/components/states/{LoadingState,EmptyState,ErrorState}.tsx`（+ `.test.tsx`、`.stories.tsx`）
- Create: `web/src/components/ui/RatingStars.tsx`（+ test、stories）
- Create: `web/src/components/pages/NotFoundPage.tsx`（+ test）
- Delete: `web/src/features/home/`、`web/src/components/pages/ComingSoonPage.tsx`

**Interfaces（PR2〜4 が使う）:**
- `LoadingState({ label?: string })` — 既定 `読み込み中…`、`role="status"`
- `EmptyState({ title: string; description?: string; action?: ReactNode })`
- `ErrorState({ title: string; description?: string; onRetry: () => void })` — `role="alert"`、ボタン名「再試行」
- `RatingStars({ rating: number; size?: 'sm' | 'lg' })` — `aria-label="評価 {rating} / 5"`
- `useGetApiBooksInfinite(params, options)` — Orval 生成（名前は生成結果に合わせ、PR 本文に書く）
- routes: `web/src/app/routes/BooksListRoute.tsx`・`BookDetailRoute.tsx`・`TagCountsRoute.tsx`（それぞれ `export function XxxRoute()`）

- [ ] **Step 1:** `Header` のテスト（ナビ「一覧」「分野別」、今いる path に `aria-current="page"`、`/books/1` ではどちらも付かない）を書いて落ちるのを見る → 実装 → 通す
- [ ] **Step 2:** `LoadingState`・`EmptyState`・`ErrorState`（再試行で `onRetry` が1回だけ呼ばれる）・`RatingStars`（1 と 5 の境界、aria-label）のテスト → 実装 → 通す
- [ ] **Step 3:** router と仮 route、`NotFoundPage`（未知の path で見つからない表示と一覧へのリンク）のテスト → 実装 → 通す。home と ComingSoon を消す
- [ ] **Step 4:** Orval の infinite 設定 → `pnpm gen:api` → 生成されたフック名を確認。Footer に楽天クレジット
- [ ] **Step 5:** 全チェックを通してコミット（指摘ごと・変更の意味ごとに分ける）、push、PR

### PR2: 一覧（`feat/public-books-list`、PR1 のマージ後）

**Files:** `web/src/features/books-list/{components,hooks,utils}/…`、`web/src/app/routes/BooksListRoute.tsx`

**Interfaces:**
- `utils/searchParams.ts`: `parseTagParam(v: string | null): number | undefined`（正の整数だけ）、`normalizeQuery(v: string): string | undefined`（trim して空なら undefined）
- `utils/paging.ts`: `nextOffset(last: BookListResponse): number | undefined`（`offset + items.length >= total` なら undefined）
- `hooks/useBooksList.ts`: URL の `q`・`tag` を読み、`useGetApiBooksInfinite({ limit: 20, q, tagId })` を呼び、`{ books, total, hasMore, loadMore, isLoadingMore, … }` を返す
- `hooks/useBookFilters.ts`: `setQuery(q)`・`setTag(id | undefined)`・`clear()`（`useSearchParams` で URL を書き換える）

- [ ] **Step 1:** utils の単体テスト（`parseTagParam`: `'3'`→3、`'0'`/`'-1'`/`'abc'`/`'1.5'`/null→undefined。`normalizeQuery`: `'  設計 '`→`'設計'`、`'   '`→undefined。`nextOffset`: offset0・20件・total45→20、offset40・5件・total45→undefined、total0→undefined）→ 実装 → 通す
- [ ] **Step 2:** 画面テスト（MSW）: 読み込み中 → 20冊表示と「N冊中 M冊を表示」／「もっと見る」で offset=20 の要求が出て最後まで読むとボタンが消える／検索の送信で `q` 付きで呼ばれ、空白だけで `q` が消える／タグで `tagId` 付き、「すべて」で外れる／タグ一覧が 500 でも本は出る／0件2種の文言と「検索と絞り込みを外す」／エラーで再試行すると取り直す／`?tag=abc` で `tagId` なし → 実装 → 通す
- [ ] **Step 3:** Presentational の stories（BookCard・BookListView の Default/Empty/Loading）
- [ ] **Step 4:** 全チェック、コミット、push、PR

### PR3: 詳細（`feat/public-book-detail`、PR1 のマージ後）

**Files:** `web/src/features/book-detail/…`、`web/src/app/routes/BookDetailRoute.tsx`

**Interfaces:**
- `utils/parseBookId.ts`: `parseBookId(v: string | undefined): number | undefined`（正の整数だけ）
- `hooks/useBookDetail.ts`: id が不正なら API を呼ばず `notFound: true`。`ApiError.status === 404` も `notFound: true`

- [ ] **Step 1:** `parseBookId` の単体テスト → 実装
- [ ] **Step 2:** 画面テスト: 全項目の表示（感想の改行、出版社なしで行が出ない、書影なしの枠）／楽天の書影のときだけクレジット・商品ページ・注記、openBD では出ない／楽天で `coverProductUrl` 空ならリンクだけ出ない／`amazonUrl` 無しで Amazon リンクなし／404 と `/books/abc` で見つからない表示（後者は API を呼ばない）／500 で再試行 → 実装 → 通す
- [ ] **Step 3:** stories（BookDetailView の openBD／楽天／書影なし）
- [ ] **Step 4:** 全チェック、コミット、push、PR

### PR4: 分野別（`feat/public-tag-counts`、PR1 のマージ後）

**Files:** `web/src/features/tag-counts/…`、`web/src/app/routes/TagCountsRoute.tsx`

**Interfaces:** `utils/barRatio.ts`: `barRatio(count: number, max: number): number`（0〜100 の整数%。max 0 なら 0）

- [ ] **Step 1:** `barRatio` の単体テスト（5/5→100、1/3→33、0/0→0）→ 実装
- [ ] **Step 2:** 画面テスト: API の順で行が並ぶ／各行が `/?tag=<id>` へのリンク／冊数の文字／0件の文言／エラーで再試行 → 実装 → 通す
- [ ] **Step 3:** stories（TagCountsView の Default/Empty）
- [ ] **Step 4:** 全チェック、コミット、push、PR
