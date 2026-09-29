import { afterEach, describe, expect, it, vi } from 'vitest'
import { adminRequest, adminToken } from './admin-auth'

describe('admin-auth', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('VITE_ADMIN_TOKEN があれば Bearer の Authorization を付ける', () => {
    vi.stubEnv('VITE_ADMIN_TOKEN', 'secret')

    expect(adminToken()).toBe('secret')
    expect(adminRequest()).toEqual({
      headers: { Authorization: 'Bearer secret' },
    })
  })

  it('前後の空白は除いて使う', () => {
    vi.stubEnv('VITE_ADMIN_TOKEN', '  secret  ')

    expect(adminToken()).toBe('secret')
  })

  it.each([
    ['空文字', ''],
    ['空白だけ', '   '],
  ])('VITE_ADMIN_TOKEN が%sなら未設定として headers を付けない', (_, value) => {
    vi.stubEnv('VITE_ADMIN_TOKEN', value)

    expect(adminToken()).toBe('')
    expect(adminRequest()).toEqual({})
  })
})
