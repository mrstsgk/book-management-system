import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import {
  getLookupCatalogMockHandler,
  getSearchTagsMockHandler,
  getRegisterBookMockHandler,
} from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useBookRegistration } from './useBookRegistration'
import { createBookFixture, createCatalogFixture } from '@/testing/fixtures'

let current: { pathname: string; state: unknown } = {
  pathname: '',
  state: null,
}
function Probe() {
  const loc = useLocation()
  current = { pathname: loc.pathname, state: loc.state }
  return null
}

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter initialEntries={['/x']}>
        <Probe />
        {children}
      </MemoryRouter>
    </QueryClientProvider>
  )
}

function renderUseBookRegistration() {
  return renderHook(() => useBookRegistration(), { wrapper })
}

describe('useBookRegistration', () => {
  it('確かめると、その ISBN の書誌と書影を返す', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '良いコード／悪いコードで学ぶ設計入門',
          coverUrl: 'https://cover.openbd.jp/9784297146221.jpg',
          coverSource: 'openbd',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
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

  it('カタログに無い ISBN は 404 として扱う', async () => {
    server.use(
      http.get('*/api/admin/catalog/:isbn', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()

    act(() => result.current.setIsbn('9784000000000'))
    act(() => result.current.confirm())

    await waitFor(() => expect(result.current.catalog.isNotFound).toBe(true))
    expect(result.current.catalog.confirmed).toBe(false)
  })

  it('登録に成功したら成功フラグを立てる', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '良いコード／悪いコードで学ぶ設計入門',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      getRegisterBookMockHandler(createBookFixture({ id: 1 })),
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

  it('登録の409（想定外の競合）は画面上部の一般エラーにし、ISBN 欄には出さない', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '本',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
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
        '登録できませんでした。時間をおいてもう一度お試しください。',
      ),
    )
    expect(result.current.isbnError).toBeUndefined()
  })

  it('登録済みの ISBN（400 の isbnDuplicated）は ISBN 欄の誤りにし、一般エラーは出さない', async () => {
    let calls = 0
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () => {
        calls += 1
        return HttpResponse.json(
          {
            message: 'invalid request',
            errors: [{ field: 'isbn', rule: 'isbnDuplicated' }],
          },
          { status: 400 },
        )
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
    act(() => result.current.submit())

    await waitFor(() =>
      expect(result.current.isbnError).toBe(
        'この ISBN の本はすでに登録されています。',
      ),
    )
    expect(result.current.generalError).toBeUndefined()
    expect(result.current.errors).toEqual({})
    expect(result.current.registered).toBe(false)
    expect(calls).toBe(1)
  })

  it('登録時の ISBN の誤りは、ISBN を書き換えたら消え、送った ISBN に戻すとまた出る', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json(
          {
            message: 'invalid request',
            errors: [{ field: 'isbn', rule: 'isbnDuplicated' }],
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
    await waitFor(() => expect(result.current.isbnError).toBeDefined())

    act(() => result.current.setIsbn('9784873118703'))
    expect(result.current.isbnError).toBeUndefined()

    act(() => result.current.setIsbn('9784297146221'))
    expect(result.current.isbnError).toBe(
      'この ISBN の本はすでに登録されています。',
    )
  })

  it('カタログに無い ISBN と無いタグの 400 は、両方を項目ごとの誤りにする', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({
        items: [{ id: 9, name: '設計', version: 1 }],
      }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json(
          {
            message: 'invalid request',
            errors: [
              { field: 'isbn', rule: 'isbnNotInCatalog' },
              { field: 'tagIds', rule: 'tagIdsNotFound' },
            ],
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
        tagIds: [9],
      }),
    )
    act(() => result.current.submit())

    await waitFor(() =>
      expect(result.current.isbnError).toBe(
        'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
      ),
    )
    expect(result.current.errors).toEqual({
      tagIds:
        '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
    })
    expect(result.current.generalError).toBeUndefined()
  })

  it('ISBN を確かめる要求が401（セッション切れ）ならログイン画面へ送る', async () => {
    server.use(
      http.get('*/api/admin/catalog/:isbn', () =>
        HttpResponse.json({ message: 'unauthorized' }, { status: 401 }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())

    await waitFor(() => expect(current.pathname).toBe('/login'))
    expect(current.state).toEqual({ from: '/x' })
  })

  it('登録が401ならログイン画面へ送る', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
        HttpResponse.json({ message: 'unauthorized' }, { status: 401 }),
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

    await waitFor(() => expect(current.pathname).toBe('/login'))
    expect(current.state).toEqual({ from: '/x' })
  })

  it('フィールドエラーの400は summary の誤りとして返す', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
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
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', () =>
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
      getLookupCatalogMockHandler(
        createCatalogFixture({ isbn: '9784297146221', title: '本' }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', async () => {
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

  it('確かめた後にISBN欄を書き換えても、確かめたISBNで登録する', async () => {
    let sentIsbn: string | undefined
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '本',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
      http.post('*/api/admin/books', async ({ request }) => {
        const body = (await request.json()) as { isbn?: string }
        sentIsbn = body.isbn
        return HttpResponse.json({ id: 1 })
      }),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))

    // 確認後に別のISBNへ書き換える（確認し直してはいない）
    act(() => result.current.setIsbn('9784000000000'))
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
    expect(sentIsbn).toBe('9784297146221')
  })

  it('必須未入力で送信後、直した項目のエラーは再送信前に消え、直していない項目は残る', async () => {
    server.use(
      getLookupCatalogMockHandler(
        createCatalogFixture({
          isbn: '9784297146221',
          title: '良いコード／悪いコードで学ぶ設計入門',
        }),
      ),
      getSearchTagsMockHandler({ items: [] }),
    )
    const { result } = renderUseBookRegistration()
    act(() => result.current.setIsbn('9784297146221'))
    act(() => result.current.confirm())
    await waitFor(() => expect(result.current.catalog.confirmed).toBe(true))

    act(() => result.current.submit())
    expect(result.current.errors.summary).toBeDefined()
    expect(result.current.errors.comment).toBeDefined()

    act(() =>
      result.current.setValues({
        ...result.current.values,
        summary: 'まとめ',
      }),
    )

    expect(result.current.errors.summary).toBeUndefined()
    expect(result.current.errors.comment).toBeDefined()
  })
})
