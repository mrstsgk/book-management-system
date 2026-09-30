import { QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import type { ReactElement } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { createQueryClient } from '@/lib/query-client'

type RouteEntry = string | { pathname: string; state?: unknown }

export function renderWithProviders(
  ui: ReactElement,
  { route = '/' }: { route?: RouteEntry } = {},
) {
  return render(
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
    </QueryClientProvider>,
  )
}
