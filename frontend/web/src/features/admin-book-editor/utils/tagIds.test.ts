import { describe, expect, it } from 'vitest'
import type { TagResponse } from '@/api/generated/api.schemas'
import { tagIdsByName } from './tagIds'

const tags: TagResponse[] = [
  { id: 1, name: '設計' },
  { id: 2, name: 'ソフトウェア設計' },
  { id: 3, name: 'データ' },
]

describe('tagIdsByName', () => {
  it('名前が完全に一致するタグの ID だけを返す', () => {
    expect(tagIdsByName(['設計', 'データ'], tags)).toEqual([1, 3])
  })

  it('別のタグ名の一部と同じでも完全一致でなければ拾わない', () => {
    expect(tagIdsByName(['ソフトウェア'], tags)).toEqual([])
  })

  it('存在しない名前は無視する', () => {
    expect(tagIdsByName(['設計', '存在しない'], tags)).toEqual([1])
  })

  it('空配列なら空を返す', () => {
    expect(tagIdsByName([], tags)).toEqual([])
  })
})
