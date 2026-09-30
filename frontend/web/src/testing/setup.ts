import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './server'

// jsdom は <dialog> の showModal / close を実装していない。開閉（open 属性）だけを再現する
// （背景を操作できなくする・フォーカスを閉じ込めるのはブラウザの仕事なので、ここでは再現しない）
if (typeof HTMLDialogElement.prototype.showModal !== 'function') {
  HTMLDialogElement.prototype.showModal = function showModal() {
    this.open = true
  }
  HTMLDialogElement.prototype.close = function close() {
    this.open = false
    this.dispatchEvent(new Event('close'))
  }
}

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => {
  // vitest の globals を使っていないため、RTL の自動 cleanup が効かない。前のテストの描画が残らないよう明示する
  cleanup()
  server.resetHandlers()
})
afterAll(() => server.close())
