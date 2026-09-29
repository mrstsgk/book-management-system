import { describe, expect, it } from 'vitest'
import type { BookListResponse } from '@/api/generated/api.schemas'
import { nextOffset } from './paging'

const page = (
  offset: number,
  count: number,
  total: number,
): BookListResponse => ({
  offset,
  limit: 20,
  total,
  items: Array.from({ length: count }, (_, i) => ({ id: offset + i + 1 })),
})

describe('nextOffset', () => {
  it('まだ続きがあれば、読み込んだ冊数の次の位置を返す', () => {
    expect(nextOffset(page(0, 20, 45))).toBe(20)
  })

  it('最後のページまで読んだら undefined を返す', () => {
    expect(nextOffset(page(40, 5, 45))).toBeUndefined()
  })

  it('ちょうど total に届いたら undefined を返す', () => {
    expect(nextOffset(page(20, 20, 40))).toBeUndefined()
  })

  it('0件なら undefined を返す', () => {
    expect(nextOffset(page(0, 0, 0))).toBeUndefined()
  })

  it('項目が空で返ったら total と食い違っても読み続けない', () => {
    expect(nextOffset(page(20, 0, 45))).toBeUndefined()
  })
})
