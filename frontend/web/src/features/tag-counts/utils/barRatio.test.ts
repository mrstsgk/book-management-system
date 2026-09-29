import { describe, expect, it } from 'vitest'
import { barRatio } from './barRatio'

describe('barRatio', () => {
  it.each([
    { count: 5, max: 5, want: 100 },
    { count: 1, max: 3, want: 33 },
    { count: 2, max: 3, want: 67 },
    { count: 0, max: 0, want: 0 },
  ])('$count / $max 冊なら棒の長さは $want%', ({ count, max, want }) => {
    expect(barRatio(count, max)).toBe(want)
  })
})
