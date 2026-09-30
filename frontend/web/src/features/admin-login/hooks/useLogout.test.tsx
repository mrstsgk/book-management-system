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

  it('失敗したらその場に留まり、失敗を伝えて、再試行できる（サーバー側のセッションが残っているため）', async () => {
    let status = 500
    server.use(
      http.post('*/api/auth/logout', () =>
        status === 204
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ message: 'x' }, { status }),
      ),
    )
    const { result } = renderHook(() => useLogout(), { wrapper })
    act(() => result.current.logout())
    await waitFor(() => expect(result.current.failed).toBe(true))
    expect(pathname).toBe('/admin/tags')

    status = 204
    act(() => result.current.logout())
    await waitFor(() => expect(pathname).toBe('/admin/login'))
    expect(result.current.failed).toBe(false)
  })
})
