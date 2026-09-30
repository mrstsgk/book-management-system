import type { TagResponse } from '@/api/generated/api.schemas'

// 本が持つ分野タグは名前で返るので、タグ一覧から ID を引き直す（完全一致だけ）
export function tagIdsByName(names: string[], tags: TagResponse[]): number[] {
  const idByName = new Map(tags.map((t) => [t.name, t.id]))
  const ids: number[] = []
  for (const name of names) {
    const id = idByName.get(name)
    if (id !== undefined) ids.push(id)
  }
  return ids
}
