import { QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useLogout } from './useLogout'

let pathname = ''
function Probe() {
  pathname = useLocation().pathname
  return null
}
function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter initialEntries={['/admin/tags']}>
        <Probe />
        <Routes>
          <Route path="*" element={children} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('useLogout', () => {
  it('POST /api/auth/logout を呼び、成功したら /admin/login へ', async () => {
    let called = false
    server.use(
      http.post('*/api/auth/logout', () => {
        called = true
        return new HttpResponse(null, { status: 204 })
      }),
    )
    const { result } = renderHook(() => useLogout(), { wrapper })
    act(() => result.current.logout())
    await waitFor(() => expect(pathname).toBe('/admin/login'))
    expect(called).toBe(true)
  })

  it('失敗してもログイン画面へ送る（Cookie はサーバーが消せなくても期限で切れる）', async () => {
    server.use(
      http.post('*/api/auth/logout', () =>
        HttpResponse.json({ message: 'x' }, { status: 500 }),
      ),
    )
    const { result } = renderHook(() => useLogout(), { wrapper })
    act(() => result.current.logout())
    await waitFor(() => expect(pathname).toBe('/admin/login'))
  })
})
