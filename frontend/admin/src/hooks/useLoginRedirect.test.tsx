import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { ApiError } from '@/api/mutator'
import { useLoginRedirect } from './useLoginRedirect'

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
    <MemoryRouter initialEntries={['/admin/tags']}>
      <Probe />
      <Routes>
        <Route path="*" element={children} />
      </Routes>
    </MemoryRouter>
  )
}

describe('useLoginRedirect', () => {
  it('401 なら元の場所を添えてログイン画面へ送り true を返す', () => {
    const { result } = renderHook(() => useLoginRedirect(), { wrapper })
    let redirected = false
    act(() => {
      redirected = result.current(new ApiError(401, 'unauthorized'))
    })
    expect(redirected).toBe(true)
    expect(current.pathname).toBe('/admin/login')
    expect(current.state).toEqual({ from: '/admin/tags' })
  })

  it.each([
    ['409', new ApiError(409, 'conflict')],
    ['通信断', new ApiError(0, 'network')],
    ['ApiError でない', new Error('x')],
  ])('%s なら何もせず false', (_label, error) => {
    const { result } = renderHook(() => useLoginRedirect(), { wrapper })
    expect(result.current(error)).toBe(false)
    expect(current.pathname).toBe('/admin/tags')
  })
})
