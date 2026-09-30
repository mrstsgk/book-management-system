import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import {
  getGetApiTagsCountsMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

function serveList(
  tags: { id: number; name: string; version: number }[],
  counts: { id: number; name: string; bookCount: number }[] = [],
) {
  server.use(
    getGetApiTagsMockHandler({ items: tags }),
    getGetApiTagsCountsMockHandler({ items: counts }),
  )
}

describe('AdminTagsRoute', () => {
  it('/admin/tags を開くと、読み込み中を出したあとタグの一覧を表示する', async () => {
    serveList(
      [{ id: 1, name: '設計', version: 1 }],
      [{ id: 1, name: '設計', bookCount: 3 }],
    )

    renderWithProviders(<AppRoutes />, { route: '/admin/tags' })

    expect(screen.getByRole('heading', { name: '分野タグ' })).toBeVisible()
    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(await screen.findByText('設計')).toBeVisible()
    expect(screen.getByText('3冊')).toBeVisible()
  })

  it('読み込みに失敗したらエラーを伝え、再試行すると取得し直す', async () => {
    let calls = 0
    server.use(
      http.get('*/api/tags', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({ items: [{ id: 1, name: '設計', version: 1 }] })
      }),
      getGetApiTagsCountsMockHandler({ items: [] }),
    )

    renderWithProviders(<AppRoutes />, { route: '/admin/tags' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'タグを読み込めませんでした',
    )

    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(await screen.findByText('設計')).toBeVisible()
    expect(calls).toBe(2)
  })

  it('タグを追加すると一覧に増える', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    server.use(
      http.post('*/api/tags', () =>
        HttpResponse.json({ id: 2, name: 'データ', version: 1 }),
      ),
    )

    renderWithProviders(<AppRoutes />, { route: '/admin/tags' })
    await screen.findByText('設計')

    serveList([
      { id: 1, name: '設計', version: 1 },
      { id: 2, name: 'データ', version: 1 },
    ])
    await userEvent.type(
      screen.getByLabelText('タグを追加', { exact: false }),
      'データ',
    )
    await userEvent.click(screen.getByRole('button', { name: '追加' }))

    expect(await screen.findByText('データ')).toBeVisible()
  })

  it('削除を確定すると一覧から消える', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    server.use(
      http.delete(
        '*/api/tags/1',
        () => new HttpResponse(null, { status: 204 }),
      ),
    )

    renderWithProviders(<AppRoutes />, { route: '/admin/tags' })
    await screen.findByText('設計')

    serveList([])
    await userEvent.click(screen.getByRole('button', { name: '削除' }))
    await userEvent.click(screen.getByRole('button', { name: '削除する' }))

    await waitFor(() =>
      expect(screen.queryByText('設計')).not.toBeInTheDocument(),
    )
  })
})
