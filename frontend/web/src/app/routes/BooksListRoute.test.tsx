import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import {
  getGetApiBooksMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import type { BookListResponse } from '@/api/generated/api.schemas'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

// total 冊ある DB を表し、受け取ったクエリを記録する
function serveBooks(total = 25) {
  const requests: URLSearchParams[] = []
  server.use(
    getGetApiBooksMockHandler(({ request }): BookListResponse => {
      const params = new URL(request.url).searchParams
      requests.push(params)
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
          authors: '著者',
          summary: '一言まとめ',
          rating: 3,
          tags: [],
        })),
      }
    }),
    getGetApiTagsMockHandler({
      items: [
        { id: 1, name: '設計' },
        { id: 2, name: 'データ' },
      ],
    }),
  )
  return requests
}

describe('BooksListRoute', () => {
  it('/ を開くと読み込み中のあと先頭の20冊と冊数を表示する', async () => {
    serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/' })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(await screen.findByText('25冊中 20冊を表示')).toBeVisible()
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(20)
  })

  it('「もっと見る」で続きを読み、最後まで読んだらボタンが消える', async () => {
    const requests = serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/' })

    await userEvent.click(
      await screen.findByRole('button', { name: 'もっと見る' }),
    )

    expect(await screen.findByText('25冊中 25冊を表示')).toBeVisible()
    expect(requests.map((r) => r.get('offset'))).toEqual(['0', '20'])
    expect(
      screen.queryByRole('button', { name: 'もっと見る' }),
    ).not.toBeInTheDocument()
  })

  it('キーワードを送信すると q 付きで読み直し、空白だけで送ると q を外す', async () => {
    const requests = serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/' })
    const input = await screen.findByRole('searchbox')

    await userEvent.type(input, ' 設計 ')
    await userEvent.click(screen.getByRole('button', { name: '検索' }))
    await waitFor(() => expect(requests.at(-1)?.get('q')).toBe('設計'))

    await userEvent.clear(screen.getByRole('searchbox'))
    await userEvent.type(screen.getByRole('searchbox'), '   ')
    await userEvent.click(screen.getByRole('button', { name: '検索' }))
    await waitFor(() => expect(requests.at(-1)?.has('q')).toBe(false))
  })

  it('タグを押すと tagId 付きで読み直し、「すべて」で絞り込みを外す', async () => {
    const requests = serveBooks()
    renderWithProviders(<AppRoutes />, { route: '/' })
    const group = await screen.findByRole('group', { name: '分野タグ' })

    await userEvent.click(within(group).getByRole('button', { name: 'データ' }))
    await waitFor(() => expect(requests.at(-1)?.get('tagId')).toBe('2'))
    expect(
      within(group).getByRole('button', { name: 'データ' }),
    ).toHaveAttribute('aria-pressed', 'true')

    await userEvent.click(within(group).getByRole('button', { name: 'すべて' }))
    await waitFor(() => expect(requests.at(-1)?.has('tagId')).toBe(false))
  })

  it.each([
    ['/?tag=1', '1'],
    ['/?tag=abc', null],
  ])(
    '%s で開くと tagId=%s で読む（分野別からのリンクは絞り込み、不正な tag は絞り込まない）',
    async (route, want) => {
      const requests = serveBooks()
      renderWithProviders(<AppRoutes />, { route })

      await waitFor(() => expect(requests).not.toHaveLength(0))
      expect(requests[0].get('tagId')).toBe(want)
    },
  )

  it('タグ一覧の取得に失敗しても本の一覧は表示する', async () => {
    serveBooks()
    server.use(
      http.get('*/api/tags', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/' })

    expect(await screen.findByText('25冊中 20冊を表示')).toBeVisible()
    expect(
      screen.queryByRole('group', { name: '分野タグ' }),
    ).not.toBeInTheDocument()
  })

  it('検索結果が0件なら条件の外し方を示し、ボタンで全件に戻す', async () => {
    const requests: URLSearchParams[] = []
    server.use(
      getGetApiBooksMockHandler(({ request }): BookListResponse => {
        const params = new URL(request.url).searchParams
        requests.push(params)
        return params.has('q')
          ? { offset: 0, limit: 20, total: 0, items: [] }
          : { offset: 0, limit: 20, total: 1, items: [{ id: 1, title: '本1' }] }
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/?q=Kubernetes' })

    await userEvent.click(
      await screen.findByRole('button', { name: '検索と絞り込みを外す' }),
    )

    expect(await screen.findByText('1冊中 1冊を表示')).toBeVisible()
    expect(requests.at(-1)?.has('q')).toBe(false)
    expect(screen.getByRole('searchbox')).toHaveValue('')
  })

  it('API がエラーなら伝え、再試行すると取得し直す', async () => {
    let calls = 0
    server.use(
      http.get('*/api/books', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({
              offset: 0,
              limit: 20,
              total: 1,
              items: [{ id: 1, title: '本1' }],
            })
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(await screen.findByText('1冊中 1冊を表示')).toBeVisible()
    expect(calls).toBe(2)
  })

  it('本とタグの両方がエラーのとき、再試行すると分野タグも復元する', async () => {
    let booksCalls = 0
    let tagsCalls = 0
    server.use(
      http.get('*/api/books', () => {
        booksCalls += 1
        return booksCalls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({
              offset: 0,
              limit: 20,
              total: 1,
              items: [{ id: 1, title: '本1' }],
            })
      }),
      http.get('*/api/tags', () => {
        tagsCalls += 1
        return tagsCalls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({ items: [{ id: 1, name: '設計' }] })
      }),
    )
    renderWithProviders(<AppRoutes />, { route: '/' })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    await userEvent.click(screen.getByRole('button', { name: '再試行' }))

    expect(await screen.findByText('1冊中 1冊を表示')).toBeVisible()
    expect(await screen.findByRole('group', { name: '分野タグ' })).toBeVisible()
    expect(tagsCalls).toBe(2)
  })
})
