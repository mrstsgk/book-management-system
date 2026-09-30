import { test, expect } from '@playwright/test'
import {
  clickPillOption,
  deleteBookById,
  deleteBookIfPresent,
  deleteTagIfPresent,
  registerBook,
  tagRow,
} from './helpers'

// Real World HTTP（openBD に書誌あり）。本のライフサイクル spec とは別の ISBN を使う。
// 名前は固定にする（実行のたびに変えると、前回の実行が途中で落ちて後片付けできなかったとき、
// 次回の beforeAll がその名前を見つけられず自己修復にならないため）
const ISBN = '9784873119038'
const TITLE = '［E2Eテスト］タグのライフサイクル'
const TAG_NAME = '［E2Eテスト］タグ'
const TAG_NAME_RENAMED = '［E2Eテスト］タグ（改名後）'

test.describe('タグのライフサイクル', () => {
  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage()
    await deleteBookIfPresent(page, TITLE)
    await deleteTagIfPresent(page, TAG_NAME)
    await deleteTagIfPresent(page, TAG_NAME_RENAMED)
    await page.close()
  })

  test('追加→本への付与→分野別集計への反映→名前変更→削除まで一気通貫で動く', async ({
    page,
  }) => {
    let bookId: string | undefined

    try {
      bookId = await registerBook(page, {
        isbn: ISBN,
        titleOverride: TITLE,
        summary: 'E2Eタグ検証用の一言まとめ',
        comment: 'E2Eタグ検証用の感想。',
        ratingLabel: '★★★',
      })

      // 追加
      await page.goto('/admin/tags')
      await page.getByLabel('タグを追加 （必須）').fill(TAG_NAME)
      await page.getByRole('button', { name: '追加' }).click()
      await expect(tagRow(page, TAG_NAME)).toBeVisible()

      // 本への付与（編集画面のチェックボックスから）
      await page.goto(`/admin/books/${bookId}/edit`)
      await clickPillOption(page, TAG_NAME)
      await page.getByRole('button', { name: '保存する' }).click()
      await expect(page).toHaveURL('/admin')

      // 分野別の集計に反映され、そこから絞り込んだ一覧へ移れる
      await page.goto('/tags')
      const tagLink = page.getByRole('link', {
        name: `${TAG_NAME} 1冊`,
        exact: true,
      })
      await expect(tagLink).toBeVisible()
      await tagLink.click()
      await expect(page).toHaveURL(/\?tag=\d+$/)
      await expect(
        page.getByRole('link', { name: TITLE, exact: false }),
      ).toBeVisible()

      // 名前変更が一覧・編集画面・公開詳細・分野別の集計すべてに反映される
      await page.goto('/admin/tags')
      await tagRow(page, TAG_NAME)
        .getByRole('button', { name: '名前を変更' })
        .click()
      await page
        .getByLabel(`「${TAG_NAME}」の新しい名前`)
        .fill(TAG_NAME_RENAMED)
      // 編集中の行は保存前まで新しい名前をまだ表示に持たない（<input> の value は
      // hasText の対象外）ため、編集を開いたときと同じ（古い名前の）行で探す
      await tagRow(page, TAG_NAME).getByRole('button', { name: '保存' }).click()
      await expect(tagRow(page, TAG_NAME_RENAMED)).toBeVisible()

      await page.goto(`/admin/books/${bookId}/edit`)
      await expect(
        page.getByRole('checkbox', { name: TAG_NAME_RENAMED, exact: true }),
      ).toBeChecked()

      await page.goto(`/books/${bookId}`)
      await expect(
        page.getByText(TAG_NAME_RENAMED, { exact: true }),
      ).toBeVisible()

      await page.goto('/tags')
      await expect(
        page.getByRole('link', {
          name: `${TAG_NAME_RENAMED} 1冊`,
          exact: true,
        }),
      ).toBeVisible()

      // 削除すると、本からもタグが外れる
      await page.goto('/admin/tags')
      await tagRow(page, TAG_NAME_RENAMED)
        .getByRole('button', { name: '削除' })
        .click()
      const dialog = page.getByRole('dialog')
      await expect(dialog).toContainText(
        '付いている1冊から外れます。本は消えません。',
      )
      await dialog.getByRole('button', { name: '削除する' }).click()
      await expect(tagRow(page, TAG_NAME_RENAMED)).toHaveCount(0)

      await page.goto(`/books/${bookId}`)
      await expect(
        page.getByText(TAG_NAME_RENAMED, { exact: true }),
      ).toHaveCount(0)
    } finally {
      // 途中で失敗しても後片付けする（タグはどちらの名前で終わったか分からないため両方試す。
      // 既に無ければ deleteTagIfPresent は何もしない）。
      // 登録自体が失敗して ID を取得できなかった場合は、書名で探して片付ける
      if (bookId !== undefined) {
        await deleteBookById(page, bookId).catch(() => {})
      } else {
        await deleteBookIfPresent(page, TITLE).catch(() => {})
      }
      await deleteTagIfPresent(page, TAG_NAME).catch(() => {})
      await deleteTagIfPresent(page, TAG_NAME_RENAMED).catch(() => {})
    }
  })
})
