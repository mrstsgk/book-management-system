import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useLogin } from './useLogin'

let pathname = ''
function Probe() {
  pathname = useLocation().pathname
  return null
}
function makeWrapper(state?: unknown) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={createQueryClient()}>
        <MemoryRouter initialEntries={[{ pathname: '/admin/login', state }]}>
          <Probe />
          <Routes>
            <Route path="*" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    )
  }
}

function serveLogin(status: number) {
  server.use(
    http.post('*/api/auth/login', () =>
      status === 204
        ? new HttpResponse(null, { status: 204 })
        : HttpResponse.json({ message: 'x' }, { status }),
    ),
  )
}

describe('useLogin', () => {
  it('空欄があれば送信しない', async () => {
    let called = false
    server.use(
      http.post('*/api/auth/login', () => {
        called = true
        return new HttpResponse(null, { status: 204 })
      }),
    )
    const { result } = renderHook(() => useLogin(), { wrapper: makeWrapper() })
    act(() => result.current.setId('admin'))
    act(() => result.current.submit())
    await waitFor(() =>
      expect(result.current.error).toBe('IDとパスワードを入力してください'),
    )
    expect(called).toBe(false)
  })

  it('成功したら from へ戻る', async () => {
    serveLogin(204)
    const { result } = renderHook(() => useLogin(), {
      wrapper: makeWrapper({ from: '/admin/tags' }),
    })
    act(() => result.current.setId('admin'))
    act(() => result.current.setPassword('pw'))
    act(() => result.current.submit())
    await waitFor(() => expect(pathname).toBe('/admin/tags'))
  })

  it('from が無ければ /admin へ', async () => {
    serveLogin(204)
    const { result } = renderHook(() => useLogin(), { wrapper: makeWrapper() })
    act(() => result.current.setId('admin'))
    act(() => result.current.setPassword('pw'))
    act(() => result.current.submit())
    await waitFor(() => expect(pathname).toBe('/admin'))
  })

  it.each([
    [401, 'IDかパスワードが違います'],
    [429, 'しばらく待ってからやり直してください'],
    [500, 'ログインできませんでした。時間をおいてもう一度お試しください。'],
  ])('%s なら「%s」', async (status, message) => {
    serveLogin(status)
    const { result } = renderHook(() => useLogin(), { wrapper: makeWrapper() })
    act(() => result.current.setId('admin'))
    act(() => result.current.setPassword('pw'))
    act(() => result.current.submit())
    await waitFor(() => expect(result.current.error).toBe(message))
    expect(pathname).toBe('/admin/login')
    // 失敗しても入力は消えない
    expect(result.current.id).toBe('admin')
  })
})
