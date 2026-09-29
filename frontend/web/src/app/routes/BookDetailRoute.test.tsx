import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'

// 仮置きの間は、path がこの画面に割り当たっていることだけを確かめる（画面ができたら差し替える）
describe('BookDetailRoute', () => {
  it('/books/1 を開くと詳細の画面（準備中）を表示し、見つからない表示にならない', () => {
    renderWithProviders(<AppRoutes />, { route: '/books/1' })

    expect(screen.getByText('準備中')).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: 'ページが見つかりませんでした' }),
    ).not.toBeInTheDocument()
  })
})
