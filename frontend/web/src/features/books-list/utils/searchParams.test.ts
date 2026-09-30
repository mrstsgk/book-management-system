import { describe, expect, it } from 'vitest'
import { normalizeQuery, parseTagParam } from './searchParams'

describe('parseTagParam', () => {
  it.each([
    ['3', 3],
    ['1', 1],
  ])('正の整数 %s はそのタグIDにする', (input, want) => {
    expect(parseTagParam(input)).toBe(want)
  })

  it('安全な整数の上限ちょうどはタグIDにする', () => {
    expect(parseTagParam(String(Number.MAX_SAFE_INTEGER))).toBe(
      Number.MAX_SAFE_INTEGER,
    )
  })

  it.each([
    ['0'],
    ['-1'],
    ['abc'],
    ['1.5'],
    [''],
    [' 3'],
    [null],
    [String(Number.MAX_SAFE_INTEGER + 1)],
    ['99999999999999999999'],
  ])('%j は絞り込みなし（undefined）にする', (input) => {
    expect(parseTagParam(input)).toBeUndefined()
  })
})

describe('normalizeQuery', () => {
  it('前後の空白を除く', () => {
    expect(normalizeQuery('  設計 ')).toBe('設計')
  })

  it.each([[''], ['   '], ['　']])(
    '空白だけの %j は undefined にする',
    (input) => {
      expect(normalizeQuery(input)).toBeUndefined()
    },
  )
})
