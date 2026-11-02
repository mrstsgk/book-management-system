import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderWithProviders } from '@/testing/render'
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
})
