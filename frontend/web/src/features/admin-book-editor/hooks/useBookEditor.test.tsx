import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import {
  getDeleteApiBooksIdMockHandler,
  getGetApiBooksIdMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useBookEditor } from './useBookEditor'

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  )
}

function renderUseBookEditor(rawId: string | undefined) {
  return renderHook(() => useBookEditor(rawId), { wrapper })
}

describe('useBookEditor', () => {
  it('本とタグ一覧が揃ったら、タグ名をIDに引き直した初期値を作る', async () => {
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '設計入門',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tags: ['設計'],
        version: 3,
      }),
      getGetApiTagsMockHandler({
        items: [
          { id: 1, name: '設計' },
          { id: 2, name: 'データ' },
        ],
      }),
    )
    const { result } = renderUseBookEditor('1')

    await waitFor(() => expect(result.current.values).not.toBeNull())
    expect(result.current.values?.tagIds).toEqual([1])
    expect(result.current.values?.summary).toBe('まとめ')
  })

  it('ID が不正なら見つからない扱いにし、API を呼ばない', () => {
    let called = false
    server.use(
      http.get('*/api/books/:id', () => {
        called = true
        return HttpResponse.json({})
      }),
    )
    const { result } = renderUseBookEditor('abc')

    expect(result.current.notFound).toBe(true)
    expect(called).toBe(false)
  })

  it('API が404なら見つからない扱いにする', async () => {
    server.use(
      http.get('*/api/books/:id', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
    )
    const { result } = renderUseBookEditor('99')

    await waitFor(() => expect(result.current.notFound).toBe(true))
  })

  it('保存に version を付け、成功したら registered を立てる', async () => {
    let sentVersion: number | undefined
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '本',
        summary: 'まとめ',
        comment: '感想',
        rating: 3,
        version: 5,
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', async ({ request }) => {
        const body = (await request.json()) as { version?: number }
        sentVersion = body.version
        return HttpResponse.json({ id: 1 })
      }),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())

    act(() => result.current.submit())

    await waitFor(() => expect(result.current.saved).toBe(true))
    expect(sentVersion).toBe(5)
  })

  it('保存が409なら競合フラグを立て、入力は消えない', async () => {
    server.use(
      getGetApiBooksIdMockHandler({
        id: 1,
        title: '本',
        summary: 'まとめ',
        comment: '感想',
        rating: 3,
        version: 5,
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())

    act(() => result.current.submit())

    await waitFor(() => expect(result.current.conflict).toBe(true))
    expect(result.current.values?.summary).toBe('まとめ')
  })

  it('最新を読み込むと、競合を解いて取り直した値になる', async () => {
    let getCalls = 0
    server.use(
      http.get('*/api/books/:id', () => {
        getCalls += 1
        return HttpResponse.json({
          id: 1,
          title: '本',
          summary: getCalls === 1 ? 'まとめ' : '取り直したまとめ',
          comment: '感想',
          rating: 3,
          version: getCalls === 1 ? 5 : 6,
        })
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())
    act(() => result.current.submit())
    await waitFor(() => expect(result.current.conflict).toBe(true))

    await act(async () => {
      await result.current.reloadLatest()
    })

    expect(result.current.values?.summary).toBe('取り直したまとめ')
    expect(result.current.conflict).toBe(false)
  })

  it('削除の確認で確定すると削除し、キャンセルでは呼ばない', async () => {
    let deleteCalled = false
    server.use(
      getGetApiBooksIdMockHandler({ id: 1, title: '本', version: 1 }),
      getGetApiTagsMockHandler({ items: [] }),
      getDeleteApiBooksIdMockHandler(() => {
        deleteCalled = true
      }),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())

    act(() => result.current.openDelete())
    expect(result.current.deleteOpen).toBe(true)

    act(() => result.current.closeDelete())
    expect(deleteCalled).toBe(false)
    expect(result.current.deleteOpen).toBe(false)

    act(() => result.current.confirmDelete())

    await waitFor(() => expect(deleteCalled).toBe(true))
  })

  it('編集中に裏で本が取り直されても、保存には編集開始時のversionを送る', async () => {
    let getCalls = 0
    let sentVersion: number | undefined
    server.use(
      http.get('*/api/books/:id', () => {
        getCalls += 1
        return HttpResponse.json({
          id: 1,
          title: '本',
          summary: 'まとめ',
          comment: '感想',
          rating: 3,
          version: getCalls === 1 ? 5 : 6,
        })
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.put('*/api/books/:id', async ({ request }) => {
        const body = (await request.json()) as { version?: number }
        sentVersion = body.version
        return HttpResponse.json({ id: 1 })
      }),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())

    // reloadLatest を経ない裏の取り直し（例: 再接続時の自動再取得）でversionだけ進む
    act(() => {
      void result.current.retry()
    })
    await waitFor(() => expect(getCalls).toBe(2))

    act(() => result.current.submit())

    await waitFor(() => expect(sentVersion).toBeDefined())
    expect(sentVersion).toBe(5)
    // フォームの値（編集開始時の内容）は書き換わっていない
    expect(result.current.values?.summary).toBe('まとめ')
  })

  it('削除に失敗したら、ダイアログにエラーを出し確定を押し直せる', async () => {
    server.use(
      getGetApiBooksIdMockHandler({ id: 1, title: '本', version: 1 }),
      getGetApiTagsMockHandler({ items: [] }),
      http.delete('*/api/books/:id', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )
    const { result } = renderUseBookEditor('1')
    await waitFor(() => expect(result.current.values).not.toBeNull())

    act(() => result.current.openDelete())
    act(() => result.current.confirmDelete())

    await waitFor(() => expect(result.current.deleteError).toBeDefined())
    expect(result.current.deleteOpen).toBe(true)
    expect(result.current.deleting).toBe(false)
  })
})
