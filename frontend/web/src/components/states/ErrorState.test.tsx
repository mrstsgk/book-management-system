import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ErrorState } from './ErrorState'

describe('ErrorState', () => {
  it('エラーを知らせ、再試行を押すと onRetry が1回だけ呼ばれる', async () => {
    const onRetry = vi.fn()
    render(
      <ErrorState
        title="本の一覧を読み込めませんでした"
        description="時間をおいてもう一度お試しください。"
        onRetry={onRetry}
      />,
    )

    expect(screen.getByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    expect(onRetry).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
