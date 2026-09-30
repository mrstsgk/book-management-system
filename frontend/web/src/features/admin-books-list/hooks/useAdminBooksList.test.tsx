import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { getGetApiBooksMockHandler } from '@/api/generated/api.msw'
import type { BookListResponse } from '@/api/generated/api.schemas'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useAdminBooksList } from './useAdminBooksList'

// total 冊ある DB を表し、受け取った offset を記録する
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
        })),
      }
    }),
  )
  return requests
}

function renderUseAdminBooksList() {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={createQueryClient()}>
      {children}
    </QueryClientProvider>
  )
  return renderHook(() => useAdminBooksList(), { wrapper })
}

describe('useAdminBooksList', () => {
  it('絞り込み無しで先頭20冊を取得し、「もっと見る」で続きを読める', async () => {
    const requests = serveBooks()
    const { result } = renderUseAdminBooksList()

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.books).toHaveLength(20)
    expect(result.current.total).toBe(25)
    expect(result.current.hasMore).toBe(true)

    await act(() => result.current.loadMore())
    await waitFor(() => expect(result.current.books).toHaveLength(25))
    expect(result.current.hasMore).toBe(false)
    expect(requests.map((r) => r.get('offset'))).toEqual(['0', '20'])
  })

  it('読み込みに失敗したら isError を返し、retry で取り直す', async () => {
    let calls = 0
    server.use(
      http.get('*/api/books', () => {
        calls += 1
        return calls === 1
          ? HttpResponse.json({ message: 'boom' }, { status: 500 })
          : HttpResponse.json({ offset: 0, limit: 20, total: 0, items: [] })
      }),
    )
    const { result } = renderUseAdminBooksList()

    await waitFor(() => expect(result.current.isError).toBe(true))
    await act(() => result.current.retry())
    await waitFor(() => expect(result.current.isError).toBe(false))
    expect(calls).toBe(2)
  })
})
