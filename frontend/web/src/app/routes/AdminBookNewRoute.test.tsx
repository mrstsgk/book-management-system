import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import {
  getLookupCatalogMockHandler,
  getSearchTagsMockHandler,
} from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'
import { createCatalogFixture } from '@/testing/fixtures'

async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/一言まとめ/), 'まとめ')
  await user.type(screen.getByLabelText(/感想/), '本文')
  await user.click(screen.getByRole('radio', { name: '★★★★' }))
}

describe('AdminBookNewRoute', () => {
  it('確かめると書誌と書影を見せる。書影が無ければ枠を出す', async () => {
    const user = userEvent.setup()
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '良いコード／悪いコードで学ぶ設計入門',
          authors: '仙塲大也',
          publisher: '技術評論社',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText('良いコード／悪いコードで学ぶ設計入門'),
    ).toBeVisible()
    expect(screen.getByText('仙塲大也')).toBeVisible()
    expect(screen.getByText('書影なし')).toBeVisible()
  })

  it('カタログに無いISBNは見つからなかったことを伝える', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('*/api/admin/catalog/:isbn', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(await screen.findByLabelText(/ISBN/), '9780000000000')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText(
        'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
      ),
    ).toBeVisible()
  })

  it('ISBNの形式が不正なら、時間をおいても解消しないことが伝わる文言を出す', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('*/api/admin/catalog/:isbn', () =>
        HttpResponse.json(
          { message: 'invalid: ISBNが不正です' },
          { status: 400 },
        ),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(await screen.findByLabelText(/ISBN/), '9784000000001')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText(
        'この ISBN は形式が正しくありません（桁数・チェックディジットを確かめてください）。',
      ),
    ).toBeVisible()
    expect(
      screen.queryByText(
        '確かめられませんでした。時間をおいてもう一度お試しください。',
      ),
    ).not.toBeInTheDocument()
  })

  it('カタログの確認が一時的な障害で失敗したら、時間をおいての再試行を促す', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('*/api/admin/catalog/:isbn', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText(
        '確かめられませんでした。時間をおいてもう一度お試しください。',
      ),
    ).toBeVisible()
  })

  it('確かめる前は登録するボタンを押せない', async () => {
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    expect(
      await screen.findByRole('button', { name: '登録する' }),
    ).toBeDisabled()
  })

  it('確かめてから登録すると成功し、一覧へ戻る', async () => {
    const user = userEvent.setup()
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '良いコード／悪いコードで学ぶ設計入門',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () => HttpResponse.json({ id: 1 })),
      // 戻り先の一覧画面（本の一覧管理）の描画を確かめるのが目的ではないので、
      // faker の乱数（評価が1〜5の範囲外になり得る）に頼らず空の一覧にする
      http.get('*/api/books', () =>
        HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('良いコード／悪いコードで学ぶ設計入門')

    await fillValidForm(user)
    await user.click(screen.getByRole('button', { name: '登録する' }))

    // /admin へ戻り、一覧（本の一覧管理画面）に完了のお知らせが出る
    expect(await screen.findByText('本を登録しました')).toBeVisible()
  })

  it('登録済みのISBN（400）は ISBN 欄の下に出し、再送信はしない', async () => {
    const user = userEvent.setup()
    let calls = 0
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () => {
        calls += 1
        return HttpResponse.json(
          {
            message: 'invalid request',
            errors: [{ field: 'isbn', rule: 'isbnDuplicated' }],
          },
          { status: 400 },
        )
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('本', { selector: 'dd' })
    await fillValidForm(user)

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(
      await screen.findByText('この ISBN の本はすでに登録されています。'),
    ).toBeVisible()
    expect(screen.getByLabelText(/ISBN/)).toHaveAttribute(
      'aria-describedby',
      'isbn-error',
    )
    expect(screen.getByLabelText(/ISBN/)).toBeInvalid()
    expect(calls).toBe(1)
  })

  it('カタログに無いISBNと無いタグの400は、両方を項目の下に出し、要約に2か所と出す', async () => {
    const user = userEvent.setup()
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({
        items: [{ id: 9, name: '設計', version: 1 }],
      }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json(
          {
            message: 'invalid request',
            errors: [
              { field: 'isbn', rule: 'isbnNotInCatalog' },
              { field: 'tagIds', rule: 'tagIdsNotFound' },
            ],
          },
          { status: 400 },
        ),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('本', { selector: 'dd' })
    await fillValidForm(user)
    await user.click(screen.getByRole('checkbox', { name: '設計' }))

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(
      await screen.findByText(
        'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
      ),
    ).toBeVisible()
    expect(
      screen.getByText(
        '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
        { selector: 'p' },
      ),
    ).toBeVisible()
    // ISBN はフォームの外の入力欄なので、要約（BookForm）は分野タグの 1 か所だけを数える
    expect(screen.getByText('1か所の入力を直してください。')).toBeVisible()
  })

  it('サーバーのフィールドエラーを項目の文言として出す', async () => {
    const user = userEvent.setup()
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json(
          {
            message: 'invalid',
            errors: [{ field: 'summary', rule: 'required' }],
          },
          { status: 400 },
        ),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('本', { selector: 'dd' })
    await fillValidForm(user)

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(
      await screen.findByText('一言まとめ: 一言まとめを入力してください'),
    ).toBeVisible()
    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ')
  })

  it('登録に失敗しても入力した値は消えない', async () => {
    const user = userEvent.setup()
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(await screen.findByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('本', { selector: 'dd' })
    await fillValidForm(user)

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(
      await screen.findByText(
        '登録できませんでした。時間をおいてもう一度お試しください。',
      ),
    ).toBeVisible()
    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ')
    expect(screen.getByLabelText(/感想/)).toHaveValue('本文')
  })
})
