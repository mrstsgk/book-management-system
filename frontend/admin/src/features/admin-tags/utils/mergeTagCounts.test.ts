import { describe, expect, it } from 'vitest'
import type {
  TagBookCountResponse,
  TagResponse,
} from '@/api/generated/api.schemas'
import { mergeTagCounts } from './mergeTagCounts'

describe('mergeTagCounts', () => {
  it('タグの一覧と冊数を id で合わせ、名前順に並べる', () => {
    const tags: TagResponse[] = [
      { id: 2, name: 'ネットワーク', version: 1 },
      { id: 1, name: 'データ', version: 3 },
    ]
    const counts: TagBookCountResponse[] = [
      { id: 2, name: 'ネットワーク', bookCount: 4 },
    ]

    expect(mergeTagCounts(tags, counts)).toEqual([
      { id: 1, name: 'データ', version: 3, bookCount: 0 },
      { id: 2, name: 'ネットワーク', version: 1, bookCount: 4 },
    ])
  })

  it('冊数の一覧に無いタグ（付いている本が0冊）は0冊として扱う', () => {
    const tags: TagResponse[] = [{ id: 1, name: '設計', version: 1 }]

    expect(mergeTagCounts(tags, [])).toEqual([
      { id: 1, name: '設計', version: 1, bookCount: 0 },
    ])
  })

  it('タグが無ければ空の一覧を返す', () => {
    expect(mergeTagCounts([], [])).toEqual([])
  })
})
