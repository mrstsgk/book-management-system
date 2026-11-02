import { describe, expect, it } from 'vitest'
import { codePointLength } from './codePointLength'

describe('codePointLength', () => {
  it.each([
    ['ASCII は 1 文字を 1 と数える', 'abc', 3],
    ['日本語は 1 文字を 1 と数える', '設計入門', 4],
    ['補助平面の絵文字（UTF-16 で 2 単位）も 1 と数える', '📚📚', 2],
    ['空文字は 0', '', 0],
  ])('%s', (_, value, expected) => {
    expect(codePointLength(value)).toBe(expected)
  })
})
