import { describe, expect, it } from 'vitest'
import { nextOffset } from './paging'

describe('nextOffset', () => {
  it.each([
    ['続きがある', { offset: 0, limit: 20, total: 45, items: Array(20) }, 20],
    [
      '最後まで読んだ',
      { offset: 40, limit: 20, total: 45, items: Array(5) },
      undefined,
    ],
    [
      '空のページ（totalと食い違っていても止める）',
      { offset: 0, limit: 20, total: 45, items: [] },
      undefined,
    ],
  ])('%s', (_name, page, want) => {
    expect(nextOffset(page)).toBe(want)
  })
})
