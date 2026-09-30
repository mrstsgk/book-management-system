import { describe, expect, it } from 'vitest'
import { emptyBookFormValues } from '../types'
import { visibleLocalErrors } from './visibleErrors'

describe('visibleLocalErrors', () => {
  it('直した項目は消え、直していない項目は残る', () => {
    const localErrors = {
      summary: '一言まとめを入力してください',
      comment: '感想を入力してください',
    }
    const values = { ...emptyBookFormValues, summary: 'まとめ' }

    const errors = visibleLocalErrors(localErrors, values)

    expect(errors.summary).toBeUndefined()
    expect(errors.comment).toBe('感想を入力してください')
  })

  it('別の理由で誤りが残るなら、今の入力に対する文言に更新される', () => {
    const localErrors = { summary: '一言まとめを入力してください' }
    const tooLong = { ...emptyBookFormValues, summary: 'a'.repeat(101) }

    expect(visibleLocalErrors(localErrors, tooLong).summary).toBe(
      '100文字以内にしてください',
    )
  })

  it('localErrors が空なら常に空', () => {
    expect(visibleLocalErrors({}, emptyBookFormValues)).toEqual({})
  })
})
