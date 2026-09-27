import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { HomePage } from './HomePage'

describe('HomePage', () => {
  it('システム名の見出しを表示する', () => {
    renderWithProviders(<HomePage />)

    expect(
      screen.getByRole('heading', { name: '書籍管理システム' }),
    ).toBeVisible()
  })
})
