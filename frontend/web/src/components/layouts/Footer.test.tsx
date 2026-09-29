import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Footer } from './Footer'

describe('Footer', () => {
  it('書誌・書影の入手元と、楽天ウェブサービスのクレジットを表示する', () => {
    render(<Footer />)

    expect(
      screen.getByText('書誌: openBD ／ 書影: openBD・楽天ブックス'),
    ).toBeVisible()
    expect(
      screen.getByRole('link', { name: 'Supported by Rakuten Developers' }),
    ).toHaveAttribute('href', 'https://webservice.rakuten.co.jp/')
  })
})
