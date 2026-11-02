import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ErrorSummary } from './ErrorSummary'

describe('ErrorSummary', () => {
  it('誤りが無ければ何も出さない', () => {
    const { container } = render(<ErrorSummary errors={[]} />)

    expect(container).toBeEmptyDOMElement()
  })

  it('誤りの件数と、各項目へのリンクを出す', () => {
    render(
      <ErrorSummary
        errors={[
          { id: 'summary', message: '一言まとめ: 入力してください' },
          { id: 'comment', message: '感想: 5000文字以内にしてください' },
        ]}
      />,
    )

    expect(screen.getByRole('alert')).toHaveTextContent(
      '2か所の入力を直してください。',
    )
    expect(
      screen.getByRole('link', { name: '一言まとめ: 入力してください' }),
    ).toHaveAttribute('href', '#summary')
    expect(
      screen.getByRole('link', { name: '感想: 5000文字以内にしてください' }),
    ).toHaveAttribute('href', '#comment')
  })
})
