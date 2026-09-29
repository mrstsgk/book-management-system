import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { RatingStars } from './RatingStars'

describe('RatingStars', () => {
  it.each([
    { rating: 1, stars: '★☆☆☆☆' },
    { rating: 3, stars: '★★★☆☆' },
    { rating: 5, stars: '★★★★★' },
  ])('評価 $rating は $stars と読み上げ用の「評価 $rating / 5」で表す', ({ rating, stars }) => {
    render(<RatingStars rating={rating} />)

    const el = screen.getByRole('img', { name: `評価 ${rating} / 5` })
    expect(el).toHaveTextContent(stars)
  })
})
