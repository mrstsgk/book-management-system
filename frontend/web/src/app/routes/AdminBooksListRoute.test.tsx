import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'

// 仮置きの間は、path が管理画面のレイアウトの中のこの画面に割り当たっていることだけを確かめる（画面ができたら差し替える）
describe('AdminBooksListRoute', () => {
  it('/admin を開くと管理画面のレイアウトで本の一覧（管理）の画面（準備中）を表示する', () => {
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(screen.getByText('準備中')).toBeVisible()
    expect(
      screen.getByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: 'ページが見つかりませんでした' }),
    ).not.toBeInTheDocument()
  })
})
