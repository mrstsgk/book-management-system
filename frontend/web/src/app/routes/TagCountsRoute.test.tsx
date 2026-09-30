import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { getGetApiTagsCountsMockHandler } from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

const counts = {
  items: [
    { id: 3, name: '設計', bookCount: 5 },
    { id: 1, name: 'クラウド', bookCount: 2 },
  ],
}

describe('TagCountsRoute', () => {
  it('/tags を開くと、読み込み中を出したあと分野ごとの冊数を絞り込んだ一覧へのリンクとして並べる', async () => {
    server.use(getGetApiTagsCountsMockHandler(counts))

    renderWithProviders(<AppRoutes />, { route: '/tags' })

    expect(screen.getByRole('heading', { name: '分野別' })).toBeVisible()
    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    const design = await screen.findByRole('link', { name: /設計/ })
    expect(design).toHaveAttribute('href', '/?tag=3')
    expect(design).toHaveTextContent('5冊')
    expect(screen.getByRole('link', { name: /クラウド/ })).toHaveAttribute(
      'href',
      '/?tag=1',
    )
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('本が付いた分野タグが無ければ、まだ無いことを伝える', async () => {
    server.use(getGetApiTagsCountsMockHandler({ items: [] }))

    renderWithProviders(<AppRoutes />, { route: '/tags' })

    expect(
      await screen.findByText('分野タグの付いた本はまだありません'),
    ).toBeVisible()
    expect(screen.queryByRole('list')).not.toBeInTheDocument()
  })

  it('読み込みに失敗したらエラーを伝え、再試行すると取得し直して表示する', async () => {
    let calls = 0
    server.use(
      http.get('*/api/tags/counts', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json(counts)
      }),
    )

    renderWithProviders(<AppRoutes />, { route: '/tags' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '分野別の冊数を読み込めませんでした',
    )

    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(await screen.findByRole('link', { name: /設計/ })).toBeVisible()
    expect(calls).toBe(2)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
