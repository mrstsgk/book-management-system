import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { createElement, type ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { getGetApiBooksIdMockHandler } from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useBookDetail } from './useBookDetail'

// hook と同名の .test.ts に置くため JSX を使わない
function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    QueryClientProvider,
    { client: createQueryClient() },
    children,
  )
}

describe('useBookDetail', () => {
  it('取得できた本を返す', async () => {
    server.use(getGetApiBooksIdMockHandler({ id: 7, title: '設計入門' }))

    const { result } = renderHook(() => useBookDetail('7'), { wrapper })

    await waitFor(() => expect(result.current.book?.title).toBe('設計入門'))
    expect(result.current.notFound).toBe(false)
    expect(result.current.isError).toBe(false)
  })

  it('ID が不正なら API を呼ばずに見つからない扱いにする', () => {
    let called = false
    server.use(
      http.get('*/api/books/:id', () => {
        called = true
        return HttpResponse.json({})
      }),
    )

    const { result } = renderHook(() => useBookDetail('abc'), { wrapper })

    expect(result.current.notFound).toBe(true)
    expect(result.current.isLoading).toBe(false)
    expect(called).toBe(false)
  })

  it('API が 404 なら見つからない扱いにし、エラー扱いにしない', async () => {
    server.use(
      http.get('*/api/books/:id', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
    )

    const { result } = renderHook(() => useBookDetail('9'), { wrapper })

    await waitFor(() => expect(result.current.notFound).toBe(true))
    expect(result.current.isError).toBe(false)
  })

  it('API が 404 以外のエラーならエラー扱いにする', async () => {
    server.use(
      http.get('*/api/books/:id', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )

    const { result } = renderHook(() => useBookDetail('9'), { wrapper })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.notFound).toBe(false)
  })
})
