import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { describedBy } from './describedBy'
import { FormField } from './FormField'

describe('FormField', () => {
  it('ラベルと必須の印を出し、入力とラベルを結ぶ', () => {
    render(
      <FormField id="summary" label="一言まとめ" required>
        <input id="summary" />
      </FormField>,
    )

    expect(screen.getByLabelText(/一言まとめ/)).toHaveAttribute('id', 'summary')
    expect(screen.getByText('（必須）')).toBeVisible()
  })

  it('必須でなければ任意の印を出す', () => {
    render(
      <FormField id="title" label="書名の上書き">
        <input id="title" />
      </FormField>,
    )

    expect(screen.getByText('（任意）')).toBeVisible()
  })

  it('ヒント・文字数・エラーを出し、describedBy の id で入力に結べる', () => {
    render(
      <FormField
        id="comment"
        label="感想"
        required
        hint="改行できます"
        error="5000文字以内にしてください"
        count={{ current: 5001, max: 5000 }}
      >
        <textarea id="comment" aria-describedby={describedBy('comment')} />
      </FormField>,
    )

    const input = screen.getByLabelText(/感想/)
    expect(input).toHaveAccessibleDescription(
      /改行できます.*5000文字以内にしてください.*5001 \/ 5000/,
    )
  })

  it('エラーが無ければエラー文言を出さない', () => {
    render(
      <FormField id="summary" label="一言まとめ" required>
        <input id="summary" aria-describedby={describedBy('summary')} />
      </FormField>,
    )

    expect(screen.getByLabelText(/一言まとめ/)).toHaveAccessibleDescription('')
  })
})
