import { describe, expect, it } from 'vitest'
import type { BookFormValues } from '../types'
import { validateBookForm } from './validateBookForm'

function values(overrides: Partial<BookFormValues> = {}): BookFormValues {
  return {
    titleOverride: '',
    summary: '一言まとめ',
    comment: '感想',
    rating: 3,
    tagIds: [],
    ...overrides,
  }
}

describe('validateBookForm', () => {
  it('すべて有効なら誤りを1つも返さない', () => {
    expect(validateBookForm(values())).toEqual({})
  })

  it('一言まとめが空なら入力してくださいと返す', () => {
    expect(validateBookForm(values({ summary: '' })).summary).toBe(
      '一言まとめを入力してください',
    )
  })

  it('一言まとめが空白だけなら未入力として扱う', () => {
    expect(validateBookForm(values({ summary: '   ' })).summary).toBe(
      '一言まとめを入力してください',
    )
  })

  it('一言まとめが100文字ちょうどなら誤りにしない', () => {
    expect(validateBookForm(values({ summary: 'あ'.repeat(100) })).summary).toBeUndefined()
  })

  it('一言まとめが101文字なら誤りにする', () => {
    expect(validateBookForm(values({ summary: 'あ'.repeat(101) })).summary).toBe(
      '100文字以内にしてください',
    )
  })

  it('感想が空なら入力してくださいと返す', () => {
    expect(validateBookForm(values({ comment: '' })).comment).toBe(
      '感想を入力してください',
    )
  })

  it('感想が5000文字ちょうどなら誤りにしない', () => {
    expect(validateBookForm(values({ comment: 'あ'.repeat(5000) })).comment).toBeUndefined()
  })

  it('感想が5001文字なら誤りにする', () => {
    expect(validateBookForm(values({ comment: 'あ'.repeat(5001) })).comment).toBe(
      '5000文字以内にしてください',
    )
  })

  it('書名の上書きが255文字ちょうどなら誤りにしない', () => {
    expect(
      validateBookForm(values({ titleOverride: 'あ'.repeat(255) })).titleOverride,
    ).toBeUndefined()
  })

  it('書名の上書きが256文字なら誤りにする', () => {
    expect(
      validateBookForm(values({ titleOverride: 'あ'.repeat(256) })).titleOverride,
    ).toBe('255文字以内にしてください')
  })

  it.each([0, 6])('評価が%iなら選んでくださいと返す', (rating) => {
    expect(validateBookForm(values({ rating })).rating).toBe('評価を選んでください')
  })

  it.each([1, 5])('評価が%iなら誤りにしない', (rating) => {
    expect(validateBookForm(values({ rating })).rating).toBeUndefined()
  })

  it('分野タグが10個ちょうどなら誤りにしない', () => {
    expect(
      validateBookForm(values({ tagIds: Array.from({ length: 10 }, (_, i) => i) }))
        .tagIds,
    ).toBeUndefined()
  })

  it('分野タグが11個なら誤りにする', () => {
    expect(
      validateBookForm(values({ tagIds: Array.from({ length: 11 }, (_, i) => i) }))
        .tagIds,
    ).toBe('分野タグは10個までにしてください')
  })
})
