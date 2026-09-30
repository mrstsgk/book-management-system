import { describe, expect, it } from 'vitest'
import { toFieldMessages } from './serverFieldErrors'

describe('toFieldMessages', () => {
  it('既知の field と rule の組み合わせを日本語の文言に変える', () => {
    expect(
      toFieldMessages([
        { field: 'summary', rule: 'required' },
        { field: 'comment', rule: 'max' },
        { field: 'rating', rule: 'min' },
        { field: 'tagIds', rule: 'max' },
        { field: 'titleOverride', rule: 'max' },
        { field: 'isbn', rule: 'required' },
      ]),
    ).toEqual({
      summary: '一言まとめを入力してください',
      comment: '5000文字以内にしてください',
      rating: '評価を選んでください',
      tagIds: '分野タグは10個までにしてください',
      titleOverride: '255文字以内にしてください',
      isbn: 'ISBNを入力してください',
    })
  })

  it('未知の rule は汎用の文言にする', () => {
    expect(toFieldMessages([{ field: 'summary', rule: 'unknown_rule' }])).toEqual({
      summary: '入力を見直してください',
    })
  })

  it('未知の field も汎用の文言にする', () => {
    expect(toFieldMessages([{ field: 'unknown_field', rule: 'required' }])).toEqual({
      unknown_field: '入力を見直してください',
    })
  })

  it('空配列なら空を返す', () => {
    expect(toFieldMessages([])).toEqual({})
  })
})
