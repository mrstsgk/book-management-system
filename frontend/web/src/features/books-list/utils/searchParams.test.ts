import { describe, expect, it } from 'vitest'
import { isQueryTooLong, normalizeQuery, parseTagParam } from './searchParams'

describe('isQueryTooLong', () => {
  // 上限は要件の 100 文字。文字数は Unicode のコードポイントで数える（補助平面の絵文字 1 つは 1 文字）
  it.each([
    ['99文字', 'あ'.repeat(99), false],
    ['100文字ちょうど', 'あ'.repeat(100), false],
    ['101文字', 'あ'.repeat(101), true],
    ['補助平面の絵文字100文字', '😀'.repeat(100), false],
    ['補助平面の絵文字101文字', '😀'.repeat(101), true],
    ['前後の空白を除くと100文字', ` ${'あ'.repeat(100)}　`, false],
    ['空', '', false],
  ])('%s なら上限超過=%s', (_name, input, want) => {
    expect(isQueryTooLong(input)).toBe(want)
  })

  it('キーワードなし（undefined）は上限超過にしない', () => {
    expect(isQueryTooLong(undefined)).toBe(false)
  })
})

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
