import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { getGetApiBooksMockHandler } from '@/api/generated/api.msw'
import type { BookListResponse } from '@/api/generated/api.schemas'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

// total 冊ある DB を表す
function serveBooks(total = 25) {
  server.use(
    getGetApiBooksMockHandler(({ request }): BookListResponse => {
      const params = new URL(request.url).searchParams
      const offset = Number(params.get('offset') ?? 0)
      const limit = Number(params.get('limit') ?? 20)
      const count = Math.max(0, Math.min(limit, total - offset))
      return {
        offset,
        limit,
        total,
        items: Array.from({ length: count }, (_, i) => ({
          id: offset + i + 1,
          title: `本${offset + i + 1}`,
          rating: 3,
          tags: [],
          summary: '一言まとめ',
        })),
      }
    }),
  )
}

describe('AdminBooksListRoute', () => {
  it('/admin を開くと読み込み中のあと先頭20冊と冊数を表を出す', async () => {
    serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(await screen.findByText('25冊中 20冊を表示')).toBeVisible()
    expect(screen.getAllByRole('row')).toHaveLength(21) // ヘッダー行 + 20冊
  })

  it('「もっと見る」で続きを読み、最後まで読んだらボタンが消える', async () => {
    serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    await userEvent.click(
      await screen.findByRole('button', { name: 'もっと見る' }),
    )

    expect(await screen.findByText('25冊中 25冊を表示')).toBeVisible()
    expect(
      screen.queryByRole('button', { name: 'もっと見る' }),
    ).not.toBeInTheDocument()
  })

  it('遷移で受け取った notice を完了のお知らせとして表示する', async () => {
    serveBooks(0)
    renderWithProviders(<AppRoutes />, {
      route: {
        pathname: '/admin',
        state: { notice: '「本1」を削除しました。' },
      },
    })

    await screen.findByRole('navigation', { name: '管理メニュー' })
    expect(await screen.findByRole('status')).toHaveTextContent(
      '「本1」を削除しました。',
    )
  })

  it('0件ならまだ本が無いことを伝える', async () => {
    serveBooks(0)
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(await screen.findByText('まだ本が登録されていません')).toBeVisible()
  })

  it('API がエラーなら伝え、再試行すると取得し直す', async () => {
    let calls = 0
    server.use(
      http.get('*/api/books', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({ offset: 0, limit: 20, total: 0, items: [] })
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(await screen.findByText('まだ本が登録されていません')).toBeVisible()
    expect(calls).toBe(2)
  })

  it('各行の編集リンクは対応する本の編集画面を指す', async () => {
    serveBooks(1)
    renderWithProviders(<AppRoutes />, { route: '/admin' })

    const row = await screen.findByRole('row', { name: /本1/ })
    expect(within(row).getByRole('link', { name: '編集' })).toHaveAttribute(
      'href',
      '/admin/books/1/edit',
    )
  })
})
