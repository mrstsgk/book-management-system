import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { getGetApiBooksIdMockHandler } from '@/api/generated/api.msw'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

describe('BookDetailRoute', () => {
  it('読み込み中を出したあと、URL の ID の本の詳細を見せる', async () => {
    let requestedId: string | undefined
    server.use(
      getGetApiBooksIdMockHandler(({ params }) => {
        requestedId = String(params.id)
        return { id: 3, title: 'ネットワークはなぜつながるのか', rating: 3 }
      }),
    )

    renderWithProviders(<AppRoutes />, { route: '/books/3' })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'ネットワークはなぜつながるのか',
      }),
    ).toBeVisible()
    expect(requestedId).toBe('3')
  })

  it('API が 404 なら本が見つからないことを伝え、一覧へ戻れる', async () => {
    server.use(
      http.get('*/api/books/:id', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
    )

    renderWithProviders(<AppRoutes />, { route: '/books/99' })

    expect(
      await screen.findByText('この本は見つかりませんでした'),
    ).toBeVisible()
    expect(screen.getByRole('link', { name: '一覧へ戻る' })).toHaveAttribute(
      'href',
      '/',
    )
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('ID が数値でなければ API を呼ばずに本が見つからないことを伝える', () => {
    let called = false
    server.use(
      http.get('*/api/books/:id', () => {
        called = true
        return HttpResponse.json({})
      }),
    )

    renderWithProviders(<AppRoutes />, { route: '/books/abc' })

    expect(screen.getByText('この本は見つかりませんでした')).toBeVisible()
    expect(called).toBe(false)
  })

  it('API がエラーならエラーを伝え、再試行すると取り直して詳細を見せる', async () => {
    let calls = 0
    server.use(
      http.get('*/api/books/:id', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({ id: 5, title: '取り直した本', rating: 5 })
      }),
    )

    renderWithProviders(<AppRoutes />, { route: '/books/5' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '本を読み込めませんでした',
    )

    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(
      await screen.findByRole('heading', { level: 1, name: '取り直した本' }),
    ).toBeVisible()
    expect(calls).toBe(2)
  })
})
