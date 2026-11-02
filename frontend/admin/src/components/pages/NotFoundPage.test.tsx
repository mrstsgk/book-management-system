import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { NotFoundPage } from './NotFoundPage'

describe('NotFoundPage', () => {
  it('ページが見つからないことを伝え、一覧へ戻るリンクを出す', () => {
    renderWithProviders(<NotFoundPage />)

    expect(
      screen.getByRole('heading', { name: 'ページが見つかりませんでした' }),
    ).toBeVisible()
    expect(screen.getByRole('link', { name: '一覧へ戻る' })).toHaveAttribute(
      'href',
      '/',
    )
  })
})
