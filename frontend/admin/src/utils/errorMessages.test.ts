import { describe, expect, it } from 'vitest'
import { getErrorMessage, toFieldMessages } from './errorMessages'

describe('toFieldMessages', () => {
  it('既知の field と rule の組み合わせを日本語の文言に変える', () => {
    expect(
      toFieldMessages([
        { field: 'summary', rule: 'required' },
        { field: 'comment', rule: 'maxLength' },
        { field: 'rating', rule: 'minimum' },
        { field: 'tagIds', rule: 'maxItems' },
        { field: 'titleOverride', rule: 'maxLength' },
        { field: 'isbn', rule: 'required' },
        { field: 'name', rule: 'maxLength' },
      ]),
    ).toEqual({
      summary: '一言まとめを入力してください',
      comment: '5000文字以内にしてください',
      rating: '評価を選んでください',
      tagIds: '分野タグは10個までにしてください',
      titleOverride: '255文字以内にしてください',
      isbn: 'ISBNを入力してください',
      name: '30文字以内にしてください',
    })
  })

  it.each([
    ['isbn', 'isbnDuplicated', 'この ISBN の本はすでに登録されています。'],
    [
      'isbn',
      'isbnNotInCatalog',
      'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
    ],
    ['isbn', 'format', 'ISBN は 13 桁か 10 桁で入力してください。'],
    [
      'tagIds',
      'tagIdsNotFound',
      '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
    ],
    [
      'name',
      'tagNameDuplicated',
      'この名前の分野タグはすでにあります。別の名前にしてください。',
    ],
  ])('相関ルール %s.%s を文言にする', (field, rule, message) => {
    expect(toFieldMessages([{ field, rule }])).toEqual({ [field]: message })
  })

  it('複数の field を同時に返す', () => {
    expect(
      toFieldMessages([
        { field: 'isbn', rule: 'isbnNotInCatalog' },
        { field: 'tagIds', rule: 'tagIdsNotFound' },
      ]),
    ).toEqual({
      isbn: 'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
      tagIds:
        '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
    })
  })

  it('未知の rule は汎用の文言にする（tagIdsDuplicated は FE から送らないので共通文言）', () => {
    expect(
      toFieldMessages([
        { field: 'summary', rule: 'unknown_rule' },
        { field: 'tagIds', rule: 'tagIdsDuplicated' },
      ]),
    ).toEqual({
      summary: '入力を見直してください',
      tagIds: '入力を見直してください',
    })
  })

  it('未知の field も汎用の文言にする', () => {
    expect(
      toFieldMessages([{ field: 'unknown_field', rule: 'required' }]),
    ).toEqual({
      unknown_field: '入力を見直してください',
    })
  })

  it('空配列なら空を返す', () => {
    expect(toFieldMessages([])).toEqual({})
  })
})

describe('getErrorMessage', () => {
  it('業務ルールのコードを文言にし、params を差し込む', () => {
    expect(
      getErrorMessage('tooManyLoginAttempts', { retryAfterSeconds: 60 }),
    ).toBe('ログインに続けて失敗したため、60秒ほど待ってからやり直してください')
  })

  it('差し込む値が足りなければ undefined（{キー} を画面に出さない）', () => {
    expect(getErrorMessage('tooManyLoginAttempts')).toBeUndefined()
  })

  it('対応の無いコードは undefined', () => {
    expect(getErrorMessage('unknownCode')).toBeUndefined()
  })

  it('field と rule の組み合わせも同じ表から引く', () => {
    expect(getErrorMessage('name.tagNameDuplicated')).toBe(
      'この名前の分野タグはすでにあります。別の名前にしてください。',
    )
  })
})
