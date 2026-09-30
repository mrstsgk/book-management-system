import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { getGetApiBooksMockHandler } from '@/api/generated/api.msw'
import type { BookListResponse } from '@/api/generated/api.schemas'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useBooksList } from './useBooksList'

// 45冊ある DB を表す。受け取った offset / limit / q / tagId を記録する
function serveBooks(total = 45) {
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
        })),
      }
    }),
  )
  return requests
}

function renderUseBooksList(route = '/') {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
    </QueryClientProvider>
  )
  return renderHook(() => useBooksList(), { wrapper })
}

describe('useBooksList', () => {
  it('URL の q と tag を条件にして先頭の20冊を読む', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList('/?q=設計&tag=3')

    await waitFor(() => expect(result.current.books).toHaveLength(20))
    expect(requests[0].get('q')).toBe('設計')
    expect(requests[0].get('tagId')).toBe('3')
    expect(requests[0].get('limit')).toBe('20')
    expect(requests[0].get('offset')).toBe('0')
    expect(result.current.total).toBe(45)
    expect(result.current.hasFilter).toBe(true)
  })

  it('tag が正の整数でなければ tagId を付けない', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList('/?tag=abc')

    await waitFor(() => expect(result.current.books).toHaveLength(20))
    expect(requests[0].has('tagId')).toBe(false)
    expect(result.current.hasFilter).toBe(false)
  })

  it('loadMore で続きを読み、最後まで読んだら hasMore が false になる', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList()

    await waitFor(() => expect(result.current.hasMore).toBe(true))
    act(() => result.current.loadMore())
    await waitFor(() => expect(result.current.books).toHaveLength(40))
    act(() => result.current.loadMore())
    await waitFor(() => expect(result.current.books).toHaveLength(45))

    expect(requests.map((r) => r.get('offset'))).toEqual(['0', '20', '40'])
    expect(result.current.hasMore).toBe(false)
  })

  it('setQuery は前後の空白を除いた q で読み直し、空白だけなら q を外す', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList('/?q=古い')

    await waitFor(() => expect(result.current.books).toHaveLength(20))
    act(() => result.current.setQuery('  新しい '))
    await waitFor(() => expect(requests).toHaveLength(2))
    expect(requests[1].get('q')).toBe('新しい')
    expect(requests[1].get('offset')).toBe('0')

    act(() => result.current.setQuery('   '))
    await waitFor(() => expect(requests).toHaveLength(3))
    expect(requests[2].has('q')).toBe(false)
  })

  it('setTag はタグで絞り込み、undefined で絞り込みを外す。q は残す', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList('/?q=設計')

    await waitFor(() => expect(result.current.books).toHaveLength(20))
    act(() => result.current.setTag(2))
    await waitFor(() => expect(requests).toHaveLength(2))
    expect(requests[1].get('tagId')).toBe('2')
    expect(requests[1].get('q')).toBe('設計')

    act(() => result.current.setTag(undefined))
    await waitFor(() => expect(requests).toHaveLength(3))
    expect(requests[2].has('tagId')).toBe(false)
    expect(requests[2].get('q')).toBe('設計')
  })

  it('clear で検索と絞り込みを両方外す', async () => {
    const requests = serveBooks()
    const { result } = renderUseBooksList('/?q=設計&tag=2')

    await waitFor(() => expect(result.current.books).toHaveLength(20))
    act(() => result.current.clear())
    await waitFor(() => expect(requests).toHaveLength(2))
    expect(requests[1].has('q')).toBe(false)
    expect(requests[1].has('tagId')).toBe(false)
  })

  it('条件を変えたら、読み込み済みの続きのページを混ぜず先頭から出し直す', async () => {
    serveBooks()
    const { result } = renderUseBooksList()

    await waitFor(() => expect(result.current.hasMore).toBe(true))
    act(() => result.current.loadMore())
    await waitFor(() => expect(result.current.books).toHaveLength(40))

    act(() => result.current.setQuery('設計'))
    await waitFor(() => expect(result.current.books).toHaveLength(20))
    expect(result.current.books[0].id).toBe(1)
  })
})
