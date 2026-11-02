import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Footer } from './Footer'

describe('Footer', () => {
  it('書誌・書影の入手元を表示する', () => {
    render(<Footer />)

    expect(screen.getByText('書誌・書影: openBD')).toBeVisible()
  })
})
