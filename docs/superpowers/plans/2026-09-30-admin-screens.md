# 管理画面（単位3） Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement each PR's task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 本の一覧・登録・編集・削除とタグの管理の画面を `frontend/web` に作る（ログインは後回し）。

**Architecture:** 土台（管理画面のレイアウト・ルーティング・トークンの渡し方・共通の部品）を PR-M0 で先に入れ、マージ後に3つの feature（`admin-books-list`・`admin-book-editor`（登録と編集・削除で同じフォーム）・`admin-tags`）を別 PR で並行して作る。feature 間 import はしない。

**Tech Stack:** React 19 / React Router 7 / TanStack Query 5（Orval 生成）/ Tailwind 3 / Vitest + RTL + MSW

**Spec:** `docs/superpowers/specs/2026-09-30-admin-screens-design.md`（見た目は `docs/design/admin-screens/*.html`）

## Global Constraints

- 生成物は手編集しない。Query フックを手書きしない（生成フックを feature の hook から呼ぶ）
- 書き込みと `GET /api/catalog` の要求には、`adminRequest()` で `Authorization` を付ける。公開画面の要求には付けない
- テスト名は日本語で振る舞いを書く。`userEvent`、`getByRole` を優先。同名のテストを置く（Stop hook）
- `cd frontend && pnpm typecheck && pnpm lint && pnpm format:check && pnpm test && pnpm build && pnpm build-storybook && pnpm gen:api:check`
- Playwright を使ったら、main checkout の `.playwright-mcp/` にできた自分のファイルを消す（並行する作業とブラウザを取り合うので、撮影はヘッドレス Chrome でもよい）
- ブランチ・コミット・PR は `docs/rules/git-*.md`。push は `gh auth status` の有効アカウントが `mrstsgk` のときだけ（資格情報を取り出して回避しない）

## Review Focus

- 一言まとめが空白だけ → 未入力として送らない（サーバーの trim と同じ扱い）
- 送信中に二度押し → 要求は1回だけ
- 保存が 409 のあと「最新を読み込む」→ 入力していた内容は最新の値で置き換わり、`version` も新しくなる（古い version で再送しない）
- 分野タグの名前が別のタグ名の一部と同じ（「設計」と「ソフトウェア設計」）→ ID の引き直しは完全一致だけ
- `VITE_ADMIN_TOKEN` が空 → 書き込みの要求を1回も送らない

---

### PR-M0: 管理画面の土台（`feat/admin-screens-foundation`、すぐ着手）

**Files:**
- Create: `web/src/lib/admin-auth.ts`（`adminToken(): string`、`adminRequest(): RequestInit`）+ test、`frontend/README.md` に `web/.env.development.local`（git 管理外）へ `VITE_ADMIN_TOKEN` を書く手順、`web/src/vite-env.d.ts` に型
- Modify: `web/src/api/mutator.ts`（`ApiError.fieldErrors`）+ test
- Create: `web/src/components/layouts/AdminLayout.tsx`（+ test。トークン未設定の注意を含む）、`web/src/components/ui/ConfirmDialog.tsx`、`web/src/components/form/{FormField,ErrorSummary}.tsx`（+ tests、stories）
- Modify: `web/src/app/router.tsx`（`/admin`・`/admin/books/new`・`/admin/books/:id/edit`・`/admin/tags` に仮置きの route。各 PR がその route ファイルだけを差し替える）、`web/src/app/routes/Admin*Route.tsx`（+ 同名の仮テスト）
- Create: `docs/adr/2026-09-30-admin-token-from-env.md`、本計画と spec（main checkout にある未コミットのファイルをコピーしてコミット）
- Modify: `docs/specifications.md` §3.1（後回し）、`frontend/README.md`（`VITE_ADMIN_TOKEN`）

**Interfaces（PR-M1〜3 が使う）:**
- `adminRequest(): RequestInit` — `{ headers: { Authorization: 'Bearer <token>' } }`。トークンが空なら headers なし
- `adminToken(): string` — 空文字なら未設定
- `ApiError.fieldErrors: { field: string; rule: string }[]`
- `ConfirmDialog({ open: boolean; title: string; description: string; confirmLabel: string; onConfirm: () => void; onCancel: () => void; busy?: boolean })`
- `FormField({ id: string; label: string; required?: boolean; hint?: string; error?: string; count?: { current: number; max: number }; children: ReactNode })` — children の入力に `aria-describedby`・`aria-invalid` を付けるのは呼ぶ側（FormField は `describedBy(id)` を返す関数も export）
- `ErrorSummary({ errors: { id: string; message: string }[] })` — 0件なら何も出さない。`role="alert"`、各項目は `#id` へのリンク
- 管理画面の完了のお知らせ: `navigate('/admin', { state: { notice: string } })`、`AdminBooksListRoute` がそれを出す
- routes: `web/src/app/routes/AdminBooksListRoute.tsx`・`AdminBookNewRoute.tsx`・`AdminBookEditRoute.tsx`・`AdminTagsRoute.tsx`

- [ ] **Step 1:** `admin-auth` と `mutator` の `fieldErrors` のテスト → 実装
- [ ] **Step 2:** `ConfirmDialog`（開閉、確定は `onConfirm` だけ・キャンセルと Esc は `onCancel` だけ、busy 中は確定を押せない）、`FormField`・`ErrorSummary` のテスト → 実装、stories
- [ ] **Step 3:** `AdminLayout`（ナビの `aria-current`、トークン未設定の注意）と router・仮 route のテスト → 実装
- [ ] **Step 4:** docs・ADR（案: A 環境変数（採用）／B ログイン画面＋sessionStorage（後回し。重い）／C mutator で全要求に付ける（公開画面の要求にトークンが載るため却下））。全チェック、コミット、push、PR

### PR-M1: 本の一覧（管理）（`feat/admin-books-list`、PR-M0 のマージ後）

**Files:** `web/src/features/admin-books-list/**`、`web/src/app/routes/AdminBooksListRoute.tsx`（+ test を書き直す）

- [ ] **Step 1:** 画面テスト: 表の各列（一言まとめが「（準備中）」なら「未記入」と出す）、「編集」のリンク先、「本を登録」のリンク、遷移の state のお知らせ（`role="status"`）が出る、「もっと見る」、0件（「まだ本が登録されていません」＋登録へのリンク）、エラー→再試行 → 実装
- [ ] **Step 2:** stories、全チェック、コミット、push、PR

### PR-M2: 本の登録・編集・削除（`feat/admin-book-editor`、PR-M0 のマージ後）

**Files:** `web/src/features/admin-book-editor/**`、`web/src/app/routes/AdminBookNewRoute.tsx`・`AdminBookEditRoute.tsx`（+ tests）

**Interfaces（feature 内）:**
- `utils/validateBookForm.ts`: `validateBookForm(v: BookFormValues): Partial<Record<keyof BookFormValues, string>>`
- `utils/serverFieldErrors.ts`: `toFieldMessages(errors: { field: string; rule: string }[]): Partial<Record<keyof BookFormValues, string>>`
- `utils/tagIds.ts`: `tagIdsByName(names: string[], tags: TagResponse[]): number[]`（完全一致だけ）

- [ ] **Step 1:** utils の単体テスト（上限ちょうど/+1、空白だけ、評価 0・6、タグ 10/11 個、未知の rule は汎用の文言、完全一致）→ 実装
- [ ] **Step 2:** 登録の画面テスト: 確かめる→書誌と書影の表示（書影が無ければ枠）、404、登録成功で `/admin` へ・お知らせ、409・カタログに無い 400・フィールドの 400 の表示、失敗しても入力が残る、二度押しで1回、Authorization が付く → 実装
- [ ] **Step 3:** 編集の画面テスト: 初期値（タグの ID 引き直し）、保存に `version`、409→「最新を読み込む」で取り直す、削除はダイアログで確定したときだけ・キャンセルでは呼ばない、404 の見つからない表示 → 実装
- [ ] **Step 4:** stories（BookForm の空・誤りあり）、全チェック、コミット、push、PR

### PR-M3: タグの管理（`feat/admin-tags`、PR-M0 のマージ後）

**Files:** `web/src/features/admin-tags/**`、`web/src/app/routes/AdminTagsRoute.tsx`（+ test）

**Interfaces（feature 内）:** `utils/mergeTagCounts.ts`: `mergeTagCounts(tags: TagResponse[], counts: TagBookCountResponse[]): { id: number; name: string; version: number; bookCount: number }[]`（名前順、counts に無いタグは 0）

- [ ] **Step 1:** `mergeTagCounts` の単体テスト → 実装
- [ ] **Step 2:** 画面テスト: 追加（成功で一覧に出る・同名 409 の文言・空白だけは送らない）、名前の変更（その場の入力・保存・キャンセル・409）、削除（ダイアログに冊数、確定したときだけ呼ぶ）、エラー→再試行、Authorization が付く → 実装
- [ ] **Step 3:** stories、全チェック、コミット、push、PR
