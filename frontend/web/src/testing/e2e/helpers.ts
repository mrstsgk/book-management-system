import { expect, type Page } from '@playwright/test'

// 本のライフサイクル・タグのライフサイクルの両 spec で共通に使う手順。

// 管理画面の操作の前に UI からログインする。資格情報は環境変数（README の手順）。
// storageState を使わないのは、spec が 2 つで UI ログインの方が短く、ログイン画面自体の配線も毎回確かめられるため
export async function login(page: Page) {
  const id = process.env.E2E_ADMIN_ID
  const password = process.env.E2E_ADMIN_PASSWORD
  if (!id || !password) {
    throw new Error(
      'E2E_ADMIN_ID と E2E_ADMIN_PASSWORD を設定してください（frontend/README.md）',
    )
  }
  await page.goto('/admin/login')
  await page.getByLabel('ID （必須）').fill(id)
  await page.getByLabel('パスワード （必須）').fill(password)
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL('/admin')
}

export type NewBookInput = {
  isbn: string
  titleOverride: string
  summary: string
  comment: string
  ratingLabel: string // 例: '★★★★'
}

// 評価・分野タグの選択肢は「見た目上の pill が <label> で、中の <input> は sr-only」という
// 構造（BookForm）。input を直接 click すると見た目上サイズが無く別要素に取られてしまうため、
// ネイティブの label→input 委譲を使い、見える文字（label のテキスト）をクリックする
export async function clickPillOption(page: Page, exactLabelText: string) {
  await page.getByText(exactLabelText, { exact: true }).click()
}

// page.goto() は画面遷移だけを待ち、画面内のAPI取得完了までは待たない（LoadingState の
// role="status" 表示が消えるまで待つ）。取得中に一覧の行数や詳細の見つからない表示を
// 判定すると、まだ何も描画されていないだけなのに「無い」と誤判定するため、
// 一覧・詳細どちらの画面に対しても goto の直後に必ず呼ぶ
async function waitForLoaded(page: Page) {
  await expect(page.getByText('読み込み中…')).toHaveCount(0)
}

// 管理画面の本の一覧で、指定した書名を含む行（無ければ空のロケーター）
export function bookRow(page: Page, titleSubstring: string) {
  return page.locator('table tbody tr', { hasText: titleSubstring })
}

function extractIdFromEditHref(href: string | null): string {
  const match = href?.match(/\/admin\/books\/(\d+)\/edit$/)
  if (!match) {
    throw new Error(`編集リンクの href から ID を取り出せませんでした: ${href}`)
  }
  return match[1]
}

// ISBN を確かめて登録する（書名の上書き・一言まとめ・感想・評価まで）。登録後の本の ID を返す
export async function registerBook(
  page: Page,
  input: NewBookInput,
): Promise<string> {
  await page.goto('/admin/books/new')
  await page.getByLabel('ISBN （必須）').fill(input.isbn)
  await page.getByRole('button', { name: '確かめる' }).click()
  // 確かめるが成功すると登録するボタンが押せるようになる（失敗のままだと押せず、ここでタイムアウトする）
  await expect(page.getByRole('button', { name: '登録する' })).toBeEnabled({
    timeout: 10_000,
  })

  await page.getByLabel('書名の上書き （任意）').fill(input.titleOverride)
  await page.getByLabel('一言まとめ （必須）').fill(input.summary)
  await page.getByLabel('感想 （必須）').fill(input.comment)
  await clickPillOption(page, input.ratingLabel)
  await page.getByRole('button', { name: '登録する' }).click()

  await expect(page).toHaveURL('/admin')
  // お知らせ（role="status"）と、一覧自体の読み込み中表示（同じく role="status"）が
  // 一瞬同居しうるため、role ではなく文言そのもので探す
  await expect(page.getByText('本を登録しました')).toBeVisible()

  const row = bookRow(page, input.titleOverride)
  await expect(row).toBeVisible()
  const href = await row
    .getByRole('link', { name: '編集' })
    .getAttribute('href')
  return extractIdFromEditHref(href)
}

// 既に削除済みの ID で呼んでも安全（何もしない）。spec の finally からは
// 「本体で既に削除済みかどうか」を追跡せずに常に呼べるようにするため
export async function deleteBookById(page: Page, bookId: string) {
  await page.goto(`/admin/books/${bookId}/edit`)
  await waitForLoaded(page)
  if (await page.getByText('この本は見つかりませんでした').isVisible()) return
  await page.getByRole('button', { name: 'この本を削除' }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: '削除する' }).click()
  await expect(page).toHaveURL('/admin')
}

// 前回の実行が途中で落ちて後片付けできなかった場合の自己修復。通常時は何もしない
export async function deleteBookIfPresent(page: Page, titleSubstring: string) {
  await page.goto('/admin')
  await waitForLoaded(page)
  const row = bookRow(page, titleSubstring)
  if ((await row.count()) === 0) return
  const href = await row
    .getByRole('link', { name: '編集' })
    .getAttribute('href')
  await deleteBookById(page, extractIdFromEditHref(href))
}

// タグの一覧（/admin/tags）で、指定した名前の行（無ければ空のロケーター）
export function tagRow(page: Page, tagName: string) {
  return page.locator('ul li', { hasText: tagName })
}

// 前回の実行が途中で落ちて後片付けできなかった場合の自己修復。通常時は何もしない
export async function deleteTagIfPresent(page: Page, tagName: string) {
  await page.goto('/admin/tags')
  await waitForLoaded(page)
  const row = tagRow(page, tagName)
  if ((await row.count()) === 0) return
  await row.getByRole('button', { name: '削除' }).click()
  await page
    .getByRole('dialog')
    .getByRole('button', { name: '削除する' })
    .click()
  await expect(tagRow(page, tagName)).toHaveCount(0)
}
