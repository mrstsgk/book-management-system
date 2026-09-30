import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useSession } from './useSession'

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      {children}
    </QueryClientProvider>
  )
}

describe('useSession', () => {
  it.each([
    [204, 'ok'],
    [401, 'unauthorized'],
    [500, 'unauthorized'],
  ])('%s なら %s（204 以外はログインし直す）', async (status, expected) => {
    server.use(
      http.get('*/api/auth/session', () =>
        status === 204
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ message: 'x' }, { status }),
      ),
    )
    const { result } = renderHook(() => useSession(), { wrapper })
    expect(result.current.status).toBe('loading')
    await waitFor(() => expect(result.current.status).toBe(expected))
  })
})
