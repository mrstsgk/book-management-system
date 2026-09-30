import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { TagResponse } from '@/api/generated/api.schemas'
import { renderWithProviders } from '@/testing/render'
import type { BookFormValues } from '../types'
import { BookForm } from './BookForm'

const tags: TagResponse[] = [
  { id: 1, name: '設計' },
  { id: 2, name: 'データ' },
]

function Wrapper({
  initial,
  ...rest
}: {
  initial: BookFormValues
  errors?: Partial<Record<keyof BookFormValues, string>>
  submitting?: boolean
  disabled?: boolean
  onSubmit?: () => void
}) {
  const [values, setValues] = useState(initial)
  return (
    <BookForm
      values={values}
      onChange={setValues}
      errors={rest.errors ?? {}}
      tags={tags}
      submitLabel="登録する"
      onSubmit={rest.onSubmit ?? (() => {})}
      submitting={rest.submitting ?? false}
      disabled={rest.disabled ?? false}
    />
  )
}

const emptyValues: BookFormValues = {
  titleOverride: '',
  summary: '',
  comment: '',
  rating: 0,
  tagIds: [],
}

describe('BookForm', () => {
  it('一言まとめ・感想を入力すると onChange に反映される', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Wrapper initial={emptyValues} />)

    await user.type(screen.getByLabelText(/一言まとめ/), 'まとめ')
    await user.type(screen.getByLabelText(/感想/), '本文')

    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ')
    expect(screen.getByLabelText(/感想/)).toHaveValue('本文')
  })

  it('評価を選ぶと選んだ星だけ選択状態になる', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Wrapper initial={emptyValues} />)

    await user.click(screen.getByRole('radio', { name: '★★★★★' }))

    expect(screen.getByRole('radio', { name: '★★★★★' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '★' })).not.toBeChecked()
  })

  it('分野タグを選ぶとチェックが付き、もう一度押すと外れる', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Wrapper initial={emptyValues} />)

    const checkbox = screen.getByRole('checkbox', { name: '設計' })
    await user.click(checkbox)
    expect(checkbox).toBeChecked()

    await user.click(checkbox)
    expect(checkbox).not.toBeChecked()
  })

  it('既存の値を初期表示する', () => {
    renderWithProviders(
      <Wrapper
        initial={{
          titleOverride: '上書き書名',
          summary: 'まとめ済み',
          comment: '感想済み',
          rating: 4,
          tagIds: [2],
        }}
      />,
    )

    expect(screen.getByLabelText(/書名の上書き/)).toHaveValue('上書き書名')
    expect(screen.getByLabelText(/一言まとめ/)).toHaveValue('まとめ済み')
    expect(screen.getByLabelText(/感想/)).toHaveValue('感想済み')
    expect(screen.getByRole('radio', { name: '★★★★' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'データ' })).toBeChecked()
  })

  it('誤りがあれば要約と各項目にエラー文言を出す', () => {
    renderWithProviders(
      <Wrapper
        initial={emptyValues}
        errors={{ summary: '一言まとめを入力してください' }}
      />,
    )

    expect(screen.getByRole('alert')).toHaveTextContent(
      '一言まとめを入力してください',
    )
    expect(screen.getByLabelText(/一言まとめ/)).toHaveAttribute(
      'aria-invalid',
      'true',
    )
  })

  it('送信中はボタンを押せない', () => {
    renderWithProviders(<Wrapper initial={emptyValues} submitting />)

    expect(screen.getByRole('button', { name: '登録する' })).toBeDisabled()
  })

  it('disabled が true ならボタンを押せない', () => {
    renderWithProviders(<Wrapper initial={emptyValues} disabled />)

    expect(screen.getByRole('button', { name: '登録する' })).toBeDisabled()
  })

  it('送信ボタンを押すと onSubmit が呼ばれる', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    renderWithProviders(<Wrapper initial={emptyValues} onSubmit={onSubmit} />)

    await user.click(screen.getByRole('button', { name: '登録する' }))

    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
