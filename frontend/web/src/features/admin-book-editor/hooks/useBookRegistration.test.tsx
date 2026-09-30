import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  getGetApiCatalogIsbnMockHandler,
  getGetApiTagsMockHandler,
  getPostApiBooksMockHandler,
} from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useBookRegistration } from './useBookRegistration'

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  )
}

function renderUseBookRegistration() {
  return renderHook(() => useBookRegistration(), { wrapper })
}

describe('useBookRegistration', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('確かめると、その ISBN の書誌と書影を返す', async () => {
    server.use(
      getGetApiCatalogIsbnMockHandler({
        isbn: '9784297146221',
        title: '良いコード／悪いコードで学ぶ設計入門',
        coverUrl: 'https://cover.openbd.jp/9784297146221.jpg',
        coverSource: 'openbd',
      }),
      getGetApiTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()

    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())

    await waitFor(() =>
      expect(result.current.catalog.data?.title).toBe(
        '良いコード／悪いコードで学ぶ設計入門',
      ),
    )
    expect(result.current.catalog.confirmed).toBe(true)
  })

  it('確かめる要求に Authorization が付く', async () => {
    vi.stubEnv('VITE_ADMIN_TOKEN', 'secret')
    let authHeader: string | null = null
    server.use(
      http.get('*/api/catalog/:isbn', ({ request }) => {
        authHeader = request.headers.get('authorization')
        return HttpResponse.json({ isbn: '9784297146221', title: '本' })
      }),
      getGetApiTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()

    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())

    await waitFor(() => expect(authHeader).toBe('Bearer secret'))
  })

  it('カタログに無い ISBN は 404 として扱う', async () => {
    server.use(
      http.get('*/api/catalog/:isbn', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
      getGetApiTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()

    act(() => result.current.setIsbn('9784000000000'))
    act(() => result.current.confirm())

    await waitFor(() => expect(result.current.catalog.isNotFound).toBe(true))
    expect(result.current.catalog.confirmed).toBe(false)
  })

  it('登録に成功したら成功フラグを立てる', async () => {
    server.use(
      getGetApiCatalogIsbnMockHandler({
        isbn: '9784297146221',
        title: '良いコード／悪いコードで学ぶ設計入門',
      }),
      getGetApiTagsMockHandler({ items: [] }),
      getPostApiBooksMockHandler({ id: 1 }),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))

    act(() =>
      result.current.setValues({
        titleOverride: '',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tagIds: [],
      }),
    )
    act(() => result.current.submit())

    await waitFor(() => expect(result.current.registered).toBe(true))
  })

  it('登録の409は「すでに登録されています」の一般エラーにする', async () => {
    server.use(
      getGetApiCatalogIsbnMockHandler({
        isbn: '9784297146221',
        title: '本',
      }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))
    act(() =>
      result.current.setValues({
        titleOverride: '',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tagIds: [],
      }),
    )
    act(() => result.current.submit())

    await waitFor(() =>
      expect(result.current.generalError).toBe(
        'この ISBN の本はすでに登録されています。',
      ),
    )
  })

  it('フィールドエラーの400は summary の誤りとして返す', async () => {
    server.use(
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
        HttpResponse.json(
          {
            message: 'invalid',
            errors: [{ field: 'summary', rule: 'required' }],
          },
          { status: 400 },
        ),
      ),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))
    act(() =>
      result.current.setValues({
        titleOverride: '',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tagIds: [],
      }),
    )
    act(() => result.current.submit())

    await waitFor(() =>
      expect(result.current.errors.summary).toBe(
        '一言まとめを入力してください',
      ),
    )
    expect(result.current.generalError).toBeUndefined()
  })

  it('失敗しても入力した値は消えない', async () => {
    server.use(
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))
    act(() =>
      result.current.setValues({
        titleOverride: '',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tagIds: [],
      }),
    )
    act(() => result.current.submit())

    await waitFor(() => expect(result.current.generalError).toBeDefined())
    expect(result.current.values.summary).toBe('まとめ')
  })

  it('二度送信しても要求は1回だけ', async () => {
    let calls = 0
    server.use(
      getGetApiCatalogIsbnMockHandler({ isbn: '9784297146221', title: '本' }),
      getGetApiTagsMockHandler({ items: [] }),
      http.post('*/api/books', async () => {
        calls += 1
        await new Promise((r) => setTimeout(r, 20))
        return HttpResponse.json({ id: 1 })
      }),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))
    act(() =>
      result.current.setValues({
        titleOverride: '',
        summary: 'まとめ',
        comment: '感想',
        rating: 4,
        tagIds: [],
      }),
    )

    act(() => {
      result.current.submit()
      result.current.submit()
    })

    await waitFor(() => expect(result.current.registered).toBe(true))
    expect(calls).toBe(1)
  })
})
