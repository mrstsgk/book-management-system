import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { EmptyState } from './EmptyState'

describe('EmptyState', () => {
  it('見出しと次にすべきことの説明、操作を表示する', () => {
    render(
      <EmptyState
        title="条件に当てはまる本はありません"
        description="キーワードを変えてください。"
        action={<button type="button">検索と絞り込みを外す</button>}
      />,
    )

    expect(screen.getByText('条件に当てはまる本はありません')).toBeVisible()
    expect(screen.getByText('キーワードを変えてください。')).toBeVisible()
    expect(
      screen.getByRole('button', { name: '検索と絞り込みを外す' }),
    ).toBeVisible()
  })
})
