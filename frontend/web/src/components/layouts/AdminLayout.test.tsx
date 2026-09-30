import { screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'

function adminNav() {
  return within(screen.getByRole('navigation', { name: '管理メニュー' }))
}

describe('AdminLayout', () => {
  it.each([
    ['/admin', '本'],
    ['/admin/books/new', '本'],
    ['/admin/books/1/edit', '本'],
    ['/admin/tags', 'タグ'],
  ])('%s では「%s」を今いる画面にする', (route, current) => {
    renderWithProviders(<AppRoutes />, { route })

    const other = current === '本' ? 'タグ' : '本'
    expect(adminNav().getByRole('link', { name: current })).toHaveAttribute(
      'aria-current',
      'page',
    )
    expect(adminNav().getByRole('link', { name: other })).not.toHaveAttribute(
      'aria-current',
    )
  })

  it('管理画面の印と、公開画面へのリンクを出し、公開画面のナビは出さない', () => {
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(adminNav().getByRole('link', { name: '本' })).toHaveAttribute(
      'href',
      '/admin',
    )
    expect(adminNav().getByRole('link', { name: 'タグ' })).toHaveAttribute(
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
})
