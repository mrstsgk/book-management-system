import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'

// 仮置きの間は、path がこの画面に割り当たっていることだけを確かめる（画面ができたら差し替える）
describe('BooksListRoute', () => {
  it('/ を開くと一覧の画面（準備中）を表示し、見つからない表示にならない', () => {
    renderWithProviders(<AppRoutes />, { route: '/' })

    expect(screen.getByText('準備中')).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: 'ページが見つかりませんでした' }),
    ).not.toBeInTheDocument()
  })
})
