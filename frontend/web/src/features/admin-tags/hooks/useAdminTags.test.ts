import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { createElement, type ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  getGetApiTagsCountsMockHandler,
  getGetApiTagsMockHandler,
} from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useAdminTags } from './useAdminTags'

// hook と同名の .test.ts に置くため JSX を使わない
function wrapper({ children }: { children: ReactNode }) {
  return createElement(
    QueryClientProvider,
    { client: createQueryClient() },
    children,
  )
}

function serveList(
  tags: { id: number; name: string; version: number }[],
  counts: { id: number; name: string; bookCount: number }[] = [],
) {
  server.use(
    getGetApiTagsMockHandler({ items: tags }),
    getGetApiTagsCountsMockHandler({ items: counts }),
  )
}

describe('useAdminTags', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_ADMIN_TOKEN', 'local-admin-token')
  })
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('タグの一覧と冊数を名前順の一覧にする', async () => {
    serveList(
      [
        { id: 2, name: 'ネットワーク', version: 1 },
        { id: 1, name: 'データ', version: 3 },
      ],
      [{ id: 2, name: 'ネットワーク', bookCount: 4 }],
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.items).toEqual([
      { id: 1, name: 'データ', version: 3, bookCount: 0 },
      { id: 2, name: 'ネットワーク', version: 1, bookCount: 4 },
    ])
  })

  it('追加に成功すると一覧を取り直す', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    let requests = 0
    let authHeader: string | null = null
    server.use(
      http.post('*/api/tags', async ({ request }) => {
        requests++
        authHeader = request.headers.get('authorization')
        return HttpResponse.json({ id: 2, name: 'データ', version: 1 })
      }),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    serveList([
      { id: 1, name: '設計', version: 1 },
      { id: 2, name: 'データ', version: 1 },
    ])
    let outcome
    await act(async () => {
      outcome = await result.current.addTag('データ')
    })

    expect(outcome).toEqual({ ok: true })
    expect(requests).toBe(1)
    expect(authHeader).toBe('Bearer local-admin-token')
    await waitFor(() =>
      expect(new Set(result.current.items.map((i) => i.name))).toEqual(
        new Set(['設計', 'データ']),
      ),
    )
  })

  it('同じ名前が既にあれば409をタグ名入りの文言にする', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    server.use(
      http.post('*/api/tags', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    let outcome
    await act(async () => {
      outcome = await result.current.addTag('設計')
    })

    expect(outcome).toEqual({
      ok: false,
      message: '「設計」というタグはすでにあります',
    })
  })

  it('名前の変更に成功すると一覧を取り直す', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    let authHeader: string | null = null
    server.use(
      http.put('*/api/tags/1', async ({ request }) => {
        authHeader = request.headers.get('authorization')
        return HttpResponse.json({
          id: 1,
          name: 'ソフトウェア設計',
          version: 2,
        })
      }),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    serveList([{ id: 1, name: 'ソフトウェア設計', version: 2 }])
    let outcome
    await act(async () => {
      outcome = await result.current.renameTag(1, 1, 'ソフトウェア設計')
    })

    expect(outcome).toEqual({ ok: true })
    expect(authHeader).toBe('Bearer local-admin-token')
    await waitFor(() =>
      expect(result.current.items[0].name).toBe('ソフトウェア設計'),
    )
  })

  it('名前の変更が409（先に更新）なら最新を読み込んで伝える', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    server.use(
      http.put('*/api/tags/1', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    serveList([{ id: 1, name: '設計（他の人が変更）', version: 2 }])
    let outcome
    await act(async () => {
      outcome = await result.current.renameTag(1, 1, '設計案')
    })

    expect(outcome).toEqual({
      ok: false,
      stale: true,
      message:
        'このタグは、ほかの画面で先に更新されていました。最新の内容を読み込みました。',
    })
    await waitFor(() =>
      expect(result.current.items[0].name).toBe('設計（他の人が変更）'),
    )
  })

  it('削除に成功すると一覧を取り直す', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    let authHeader: string | null = null
    server.use(
      http.delete('*/api/tags/1', async ({ request }) => {
        authHeader = request.headers.get('authorization')
        return new HttpResponse(null, { status: 204 })
      }),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    serveList([])
    let outcome
    await act(async () => {
      outcome = await result.current.deleteTag(1)
    })

    expect(outcome).toEqual({ ok: true })
    expect(authHeader).toBe('Bearer local-admin-token')
    await waitFor(() => expect(result.current.items).toEqual([]))
  })

  it('削除に失敗すれば分かる文言を返す', async () => {
    serveList([{ id: 1, name: '設計', version: 1 }])
    server.use(
      http.delete('*/api/tags/1', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )

    const { result } = renderHook(() => useAdminTags(), { wrapper })
    await waitFor(() => expect(result.current.isPending).toBe(false))

    let outcome
    await act(async () => {
      outcome = await result.current.deleteTag(1)
    })

    expect(outcome).toEqual({
      ok: false,
      message: '削除できませんでした。時間をおいてもう一度お試しください。',
    })
  })
})
