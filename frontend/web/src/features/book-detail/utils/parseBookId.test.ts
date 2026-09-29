import { describe, expect, it } from 'vitest'
import { parseBookId } from './parseBookId'

describe('parseBookId', () => {
  it.each([
    ['1', 1],
    ['42', 42],
  ])('%s は正の整数 %d として読む', (raw, want) => {
    expect(parseBookId(raw)).toBe(want)
  })

  it.each([['0'], ['-1'], ['abc'], ['1.5'], ['1abc'], [''], [undefined]])(
    '%s は本の ID として扱わない（undefined）',
    (raw) => {
      expect(parseBookId(raw)).toBeUndefined()
    },
  )
})
