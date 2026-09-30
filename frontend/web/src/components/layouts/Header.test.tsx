import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { Header } from './Header'

describe('Header', () => {
  it('サイト名は一覧へのリンクになる', () => {
    renderWithProviders(<Header />, { route: '/tags' })

    expect(screen.getByRole('link', { name: '読んだ本' })).toHaveAttribute(
      'href',
      '/',
    )
  })

  it.each([
    { route: '/', current: '一覧', other: '分野別' },
    { route: '/tags', current: '分野別', other: '一覧' },
  ])(
    '$route では「$current」だけが今いる画面になる',
    ({ route, current, other }) => {
      renderWithProviders(<Header />, { route })

      expect(screen.getByRole('link', { name: current })).toHaveAttribute(
        'aria-current',
        'page',
      )
      expect(screen.getByRole('link', { name: other })).not.toHaveAttribute(
        'aria-current',
      )
    },
  )

  it('詳細画面ではナビのどちらも今いる画面にならない', () => {
    renderWithProviders(<Header />, { route: '/books/1' })

    for (const name of ['一覧', '分野別']) {
      expect(screen.getByRole('link', { name })).not.toHaveAttribute(
        'aria-current',
      )
    }
  })
})
