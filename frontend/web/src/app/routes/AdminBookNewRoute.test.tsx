import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  getGetApiCatalogIsbnMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/一言まとめ/), 'まとめ')
  await user.type(screen.getByLabelText(/感想/), '本文')
  await user.click(screen.getByRole('radio', { name: '★★★★' }))
}

describe('AdminBookNewRoute', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('確かめると書誌と書影を見せる。書影が無ければ枠を出す', async () => {
    const user = userEvent.setup()
    server.use(
      getGetApiCatalogIsbnMockHandler({
        isbn: '9784297146221',
        title: '良いコード／悪いコードで学ぶ設計入門',
        authors: '仙塲大也',
        publisher: '技術評論社',
      }),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
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
      http.get('*/api/catalog/:isbn', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(screen.getByLabelText(/ISBN/), '9780000000000')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText(
        'この ISBN の本は外部カタログに見つかりませんでした。',
      ),
    ).toBeVisible()
  })

  it('ISBNの形式が不正なら、時間をおいても解消しないことが伝わる文言を出す', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('*/api/catalog/:isbn', () =>
        HttpResponse.json(
          { message: 'invalid: ISBNが不正です' },
          { status: 400 },
        ),
      ),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(screen.getByLabelText(/ISBN/), '9784000000001')
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
      http.get('*/api/catalog/:isbn', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    expect(
      await screen.findByText(
        '確かめられませんでした。時間をおいてもう一度お試しください。',
      ),
    ).toBeVisible()
  })

  it('確かめる前は登録するボタンを押せない', () => {
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    expect(screen.getByRole('button', { name: '登録する' })).toBeDisabled()
  })

  it('確かめてから登録すると成功し、一覧へ戻る', async () => {
    const user = userEvent.setup()
    server.use(
      getGetApiCatalogIsbnMockHandler({
        isbn: '9784297146221',
        title: '良いコード／悪いコードで学ぶ設計入門',
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () => HttpResponse.json({ id: 1 })),
      // 戻り先の一覧画面（本の一覧管理）の描画を確かめるのが目的ではないので、
      // faker の乱数（評価が1〜5の範囲外になり得る）に頼らず空の一覧にする
      http.get('*/api/books', () =>
        HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('良いコード／悪いコードで学ぶ設計入門')

    await fillValidForm(user)
    await user.click(screen.getByRole('button', { name: '登録する' }))

    // /admin へ戻り、一覧（本の一覧管理画面）に完了のお知らせが出る
    expect(await screen.findByText('本を登録しました')).toBeVisible()
  })

  it('同じISBNの本が既にあれば409の文言を出す', async () => {
    const user = userEvent.setup()
    server.use(
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))
    await screen.findByText('本', { selector: 'dd' })
    await fillValidForm(user)

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(
      await screen.findByText('この ISBN の本はすでに登録されています。'),
    ).toBeVisible()
  })

  it('サーバーのフィールドエラーを項目の文言として出す', async () => {
    const user = userEvent.setup()
    server.use(
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
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
    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
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
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })
    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
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

  it('要求に Authorization が付く', async () => {
    vi.stubEnv('VITE_ADMIN_TOKEN', 'secret')
    const user = userEvent.setup()
    let authHeader: string | null = null
    server.use(
      http.get('*/api/catalog/:isbn', ({ request }) => {
        authHeader = request.headers.get('authorization')
        return HttpResponse.json({ isbn: '9784297146221', title: '本' })
      }),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/new' })

    await user.type(screen.getByLabelText(/ISBN/), '9784297146221')
    await user.click(screen.getByRole('button', { name: '確かめる' }))

    await screen.findByText('本', { selector: 'dd' })
    expect(authHeader).toBe('Bearer secret')
  })
})
