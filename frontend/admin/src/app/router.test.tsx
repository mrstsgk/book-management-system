import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'
import { AppRoutes } from './router'

describe('AppRoutes', () => {
  it('管理画面の知らない path では、管理画面のレイアウトのままページが見つからないことを伝える', async () => {
    renderWithProviders(<AppRoutes />, { route: '/unknown' })

    expect(
      await screen.findByRole('heading', {
        name: 'ページが見つかりませんでした',
      }),
    ).toBeVisible()
    expect(
      screen.getByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
  })

  it('未ログインで /login を開くと管理画面のレイアウトの外でログイン画面を出す', async () => {
    server.use(
      http.get('*/api/admin/auth/session', () =>
        HttpResponse.json({ message: 'x' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/login' })

    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('navigation', { name: '管理メニュー' }),
    ).not.toBeInTheDocument()
  })

  it('未ログインで /books を開くとログイン画面へ送る', async () => {
    server.use(
      http.get('*/api/admin/auth/session', () =>
        HttpResponse.json({ message: 'x' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/books' })

    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
  })
})
