import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { getCountBooksMockHandler } from '@/api/generated/api.msw'
import { createQueryClient } from '@/lib/query-client'
import { server } from '@/testing/server'
import { useTagCounts } from './useTagCounts'

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createQueryClient()}>
      {children}
    </QueryClientProvider>
  )
}

describe('useTagCounts', () => {
  it('API が返した分野と冊数を、その順のまま返す', async () => {
    server.use(
      getCountBooksMockHandler({
        items: [
          { id: 2, name: '設計', bookCount: 4 },
          { id: 1, name: 'データ', bookCount: 1 },
        ],
      }),
    )

    const { result } = renderHook(() => useTagCounts(), { wrapper })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.items.map((item) => item.name)).toEqual([
      '設計',
      'データ',
    ])
  })
})
