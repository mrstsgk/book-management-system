import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { getGetApiBooksMockHandler } from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

async function findAdminNav() {
  return await screen.findByRole('navigation', { name: '管理メニュー' })
}

describe('AdminLayout', () => {
  // /admin の本一覧は faker の乱数データに頼らず、空の一覧に固定する
  beforeEach(() => {
    server.use(
      getGetApiBooksMockHandler({ items: [], total: 0, limit: 20, offset: 0 }),
    )
  })

  it.each([
    ['/admin', '本'],
    ['/admin/books/new', '本'],
    ['/admin/books/1/edit', '本'],
    ['/admin/tags', 'タグ'],
  ])('%s では「%s」を今いる画面にする', async (route, current) => {
    renderWithProviders(<AppRoutes />, { route })

    const other = current === '本' ? 'タグ' : '本'
    const nav = within(await findAdminNav())
    expect(nav.getByRole('link', { name: current })).toHaveAttribute(
      'aria-current',
      'page',
    )
    expect(nav.getByRole('link', { name: other })).not.toHaveAttribute(
      'aria-current',
    )
  })

  it('管理画面の印と、公開画面へのリンクを出し、公開画面のナビは出さない', async () => {
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    const nav = within(await findAdminNav())
    expect(nav.getByRole('link', { name: '本' })).toHaveAttribute(
      'href',
      '/admin',
    )
    expect(nav.getByRole('link', { name: 'タグ' })).toHaveAttribute(
      'href',
      '/admin/tags',
    )
    expect(screen.getByText('管理')).toBeVisible()
    expect(
      screen.getByRole('link', { name: '公開画面を見る' }),
    ).toHaveAttribute('href', '/')
    expect(
      screen.queryByRole('navigation', { name: 'メイン' }),
    ).not.toBeInTheDocument()
  })

  it('ログアウトを押すと POST /api/auth/logout を呼び、ログイン画面へ移る', async () => {
    let called = false
    server.use(
      http.post('*/api/auth/logout', () => {
        called = true
        return new HttpResponse(null, { status: 204 })
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin' })
    await userEvent.click(
      await screen.findByRole('button', { name: 'ログアウト' }),
    )
    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    expect(called).toBe(true)
  })

  it('ログアウトに失敗したら管理画面に留まり、失敗を伝える（サーバー側のセッションが残っているため）', async () => {
    server.use(
      http.post('*/api/auth/logout', () =>
        HttpResponse.json({ message: 'internal' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin' })
    await userEvent.click(
      await screen.findByRole('button', { name: 'ログアウト' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'ログアウトできませんでした',
    )
    expect(
      within(await findAdminNav()).getByRole('link', { name: '本' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: '管理画面にログイン' }),
    ).not.toBeInTheDocument()
  })
})
