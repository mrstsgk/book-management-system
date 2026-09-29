import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './server'

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => {
  // vitest の globals を使っていないため、RTL の自動 cleanup が効かない。前のテストの描画が残らないよう明示する
  cleanup()
  server.resetHandlers()
})
afterAll(() => server.close())
