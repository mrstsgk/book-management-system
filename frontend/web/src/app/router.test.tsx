import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'
import { AppRoutes } from './router'

describe('AppRoutes', () => {
  it('知らない path ではページが見つからないことを伝え、一覧へ戻れる', () => {
    renderWithProviders(<AppRoutes />, { route: '/unknown' })

    expect(
      screen.getByRole('heading', { name: 'ページが見つかりませんでした' }),
    ).toBeVisible()
    expect(screen.getByRole('link', { name: '一覧へ戻る' })).toHaveAttribute(
      'href',
      '/',
    )
  })

  it('管理画面の知らない path では、管理画面のレイアウトのままページが見つからないことを伝える', async () => {
    renderWithProviders(<AppRoutes />, { route: '/admin/unknown' })

    expect(
      await screen.findByRole('heading', {
        name: 'ページが見つかりませんでした',
      }),
    ).toBeVisible()
    expect(
      screen.getByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
  })

  it('/admin/login は管理画面のレイアウトの外でログイン画面を出す', () => {
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(
      screen.getByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('navigation', { name: '管理メニュー' }),
    ).not.toBeInTheDocument()
  })

  it('未ログインで /admin を開くとログイン画面へ送る', async () => {
    server.use(
      http.get('*/api/auth/session', () =>
        HttpResponse.json({ message: 'x' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
  })
})
