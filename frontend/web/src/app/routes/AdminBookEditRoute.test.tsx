import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  getDeleteApiBooksIdMockHandler,
  getGetApiBooksIdMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

describe('AdminBookEditRoute', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('本とタグを読み込んで初期値を表示する（タグは名前からIDに引き直す）', async () => {
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '設計入門',
        isbn: '9784297146221',
        summary: 'まとめ',
        comment: '感想の本文',
        rating: 4,
        tags: ['設計'],
        version: 3,
      }),
      getGetApiTagsMockHandler({
        items: [
          { id: 1, name: '設計' },
          { id: 2, name: 'データ' },
        ],
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/1/edit' })

    expect(
      await screen.findByRole('heading', { level: 1, name: '本を編集' }),
    ).toBeVisible()
    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ')
    expect(screen.getByLabelText(/感想/)).toHaveValue('感想の本文')
    expect(screen.getByRole('radio', { name: '★★★★' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: '設計' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'データ' })).not.toBeChecked()
  })

  it('存在しない本なら見つからないことを伝える', async () => {
    server.use(
      http.get('*/api/books/:id', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/99/edit' })

    expect(
      await screen.findByText('この本は見つかりませんでした'),
    ).toBeVisible()
  })

  it('保存すると一覧へ戻る', async () => {
    const user = userEvent.setup()
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '本',
        summary: 'まとめ',
        comment: '感想',
        rating: 3,
        version: 5,
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', () => HttpResponse.json({ id: 1 })),
      // 戻り先の一覧画面（本の一覧管理）の描画を確かめるのが目的ではないので、
      // faker の乱数（評価が1〜5の範囲外になり得る）に頼らず空の一覧にする
      http.get('*/api/books', () =>
        HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/1/edit' })
    await screen.findByRole('heading', { level: 1, name: '本を編集' })

    await user.click(screen.getByRole('button', { name: '保存する' }))

    expect(await screen.findByText('本を更新しました')).toBeVisible()
  })

  it('保存が先に更新されていたら注意を出し、最新を読み込むと入れ直せる', async () => {
    const user = userEvent.setup()
    let getCalls = 0
    server.use(
      http.get('*/api/books/:id', () => {
        getCalls += 1
        return HttpResponse.json({
          id: 1,
          title: '本',
          summary: getCalls === 1 ? 'まとめ' : '取り直したまとめ',
          comment: '感想',
          rating: 3,
          version: getCalls === 1 ? 5 : 6,
        })
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/1/edit' })
    await screen.findByRole('heading', { level: 1, name: '本を編集' })

    await user.click(screen.getByRole('button', { name: '保存する' }))

    expect(
      await screen.findByText(
        'この本は、ほかの画面で先に更新されていたため保存しませんでした。',
      ),
    ).toBeVisible()
    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ')

    await user.click(screen.getByRole('button', { name: '最新を読み込む' }))

    expect(await screen.findByLabelText(/一言まとめ/)).toHaveValue(
      '取り直したまとめ',
    )
    expect(
      screen.queryByText(
        'この本は、ほかの画面で先に更新されていたため保存しませんでした。',
      ),
    ).not.toBeInTheDocument()
  })

  it('削除は確認してから行う。キャンセルでは削除しない', async () => {
    const user = userEvent.setup()
    let deleteCalled = false
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '削除対象の本',
        version: 1,
      }),
      getGetApiTagsMockHandler({ items: [] }),
      getDeleteApiBooksIdMockHandler(() => {
        deleteCalled = true
      }),
      // 戻り先の一覧画面（本の一覧管理）の描画を確かめるのが目的ではないので、
      // faker の乱数（評価が1〜5の範囲外になり得る）に頼らず空の一覧にする
      http.get('*/api/books', () =>
        HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/1/edit' })
    await screen.findByRole('heading', { level: 1, name: '本を編集' })

    await user.click(screen.getByRole('button', { name: 'この本を削除' }))
    expect(
      screen.getByRole('heading', { name: '「削除対象の本」を削除しますか？' }),
    ).toBeVisible()

    await user.click(screen.getByRole('button', { name: 'キャンセル' }))
    expect(deleteCalled).toBe(false)

    await user.click(screen.getByRole('button', { name: 'この本を削除' }))
    await user.click(screen.getByRole('button', { name: '削除する' }))

    expect(
      await screen.findByText('『削除対象の本』を削除しました'),
    ).toBeVisible()
    expect(deleteCalled).toBe(true)
  })

  it('要求に Authorization が付く', async () => {
    vi.stubEnv('VITE_ADMIN_TOKEN', 'secret')
    let authHeader: string | null = null
    server.use(
      http.get('*/api/books/:id', ({ request }) => {
        authHeader = request.headers.get('authorization')
        return HttpResponse.json({ id: 1, title: '本', version: 1 })
      }),
      getGetApiTagsMockHandler({ items: [] }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/books/1/edit' })

    await screen.findByRole('heading', { level: 1, name: '本を編集' })
    expect(authHeader).toBe('Bearer secret')
  })
})
