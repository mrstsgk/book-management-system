import { test, expect } from '@playwright/test'
import {
  bookRow,
  deleteBookById,
  deleteBookIfPresent,
  registerBook,
} from './helpers'

// リーダブルコード（openBD に書誌あり）。見本データ・他 spec のタグライフサイクルとは
// 別の ISBN を使い、テスト同士のデータが重ならないようにする。書名は実行のたびに
// 変えない（固定にしないと、前回の実行が途中で落ちて後片付けできなかったとき、
// 次回の beforeAll がその名前を見つけられず自己修復にならないため）
const ISBN = '9784873115658'
const TITLE = '［E2Eテスト］本のライフサイクル'

test.describe('本のライフサイクル', () => {
  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage()
    await deleteBookIfPresent(page, TITLE)
    await page.close()
  })

  test('登録が公開一覧・詳細に反映され、編集・削除まで一気通貫で動く', async ({
    page,
  }) => {
    let bookId: string | undefined

    try {
      bookId = await registerBook(page, {
        isbn: ISBN,
        titleOverride: TITLE,
        summary: 'E2Eで登録した検証用の一言まとめ',
        comment: 'E2Eで登録した検証用の感想。',
        ratingLabel: '★★★★',
      })

      // 公開一覧: キーワード検索で見つかり、詳細へ遷移できる
      await page.goto('/')
      await page
        .getByRole('searchbox', { name: 'キーワード（書名・著者）' })
        .fill(TITLE)
      await page.getByRole('button', { name: '検索' }).click()
      const listItem = page.getByRole('link', { name: TITLE, exact: false })
      await expect(listItem).toBeVisible()
      await listItem.click()
      await expect(page).toHaveURL(`/books/${bookId}`)

      // 公開詳細: 登録した内容がすべて反映されている
      await expect(
        page.getByRole('heading', { name: TITLE, level: 1 }),
      ).toBeVisible()
      await expect(
        page.getByText('E2Eで登録した検証用の一言まとめ'),
      ).toBeVisible()
      await expect(page.getByText('E2Eで登録した検証用の感想。')).toBeVisible()
      await expect(page.getByRole('img', { name: '評価 4 / 5' })).toBeVisible()
      await expect(
        page.getByRole('link', { name: 'Amazonで見る' }),
      ).toHaveAttribute('href', /amazon\.co\.jp/)

      // 編集: 一言まとめを変えて保存すると、公開詳細にも反映される
      await page.goto(`/admin/books/${bookId}/edit`)
      await page
        .getByLabel('一言まとめ （必須）')
        .fill('E2Eで編集した一言まとめ')
      await page.getByRole('button', { name: '保存する' }).click()
      await expect(page).toHaveURL('/admin')
      await expect(page.getByText('本を更新しました')).toBeVisible()

      await page.goto(`/books/${bookId}`)
      await expect(page.getByText('E2Eで編集した一言まとめ')).toBeVisible()

      // 削除: 確認してから削除し、管理一覧・公開画面の両方から消える
      await page.goto(`/admin/books/${bookId}/edit`)
      await page.getByRole('button', { name: 'この本を削除' }).click()
      const dialog = page.getByRole('dialog')
      await expect(dialog).toBeVisible()
      await dialog.getByRole('button', { name: '削除する' }).click()

      await expect(page).toHaveURL('/admin')
      await expect(page.getByText(`『${TITLE}』を削除しました`)).toBeVisible()
      await expect(bookRow(page, TITLE)).toHaveCount(0)

      await page.goto(`/books/${bookId}`)
      await expect(page.getByText('この本は見つかりませんでした')).toBeVisible()
    } finally {
      // 途中で失敗しても後片付けする（削除まで成功していれば 404 になるだけなので無視）。
      // 登録自体が失敗して ID を取得できなかった場合は、書名で探して片付ける
      if (bookId !== undefined) {
        await deleteBookById(page, bookId).catch(() => {})
      } else {
        await deleteBookIfPresent(page, TITLE).catch(() => {})
      }
    }
  })
})
