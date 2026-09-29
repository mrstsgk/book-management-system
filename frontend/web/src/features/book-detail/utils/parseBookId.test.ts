import { describe, expect, it } from 'vitest'
import { parseBookId } from './parseBookId'

describe('parseBookId', () => {
  it.each([
    ['1', 1],
    ['42', 42],
  ])('%s は正の整数 %d として読む', (raw, want) => {
    expect(parseBookId(raw)).toBe(want)
  })

  it('安全に扱える最大の整数までは ID として読む', () => {
    expect(parseBookId(String(Number.MAX_SAFE_INTEGER))).toBe(
      Number.MAX_SAFE_INTEGER,
    )
  })

  it.each([
    ['0'],
    ['-1'],
    ['abc'],
    ['1.5'],
    ['1abc'],
    [''],
    [undefined],
    [String(Number.MAX_SAFE_INTEGER + 1)],
    ['99999999999999999999'],
  ])('%s は本の ID として扱わない（undefined）', (raw) => {
    expect(parseBookId(raw)).toBeUndefined()
  })
})
