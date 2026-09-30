import type { BookListResponse } from '@/api/generated/api.schemas'

export const PAGE_SIZE = 20

// 空のページが返ったら total と食い違っていても止める（同じ offset を読み続けないため）
export function nextOffset(last: BookListResponse): number | undefined {
  const count = last.items?.length ?? 0
  const next = (last.offset ?? 0) + count
  if (count === 0 || next >= (last.total ?? 0)) return undefined
  return next
}
