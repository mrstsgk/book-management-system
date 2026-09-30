import type {
  TagBookCountResponse,
  TagResponse,
} from '@/api/generated/api.schemas'

export type TagWithCount = {
  id: number
  name: string
  version: number
  bookCount: number
}

// タグの一覧（全件）と冊数の一覧（1冊以上のタグだけ）を id で合わせ、名前順に並べる
export function mergeTagCounts(
  tags: TagResponse[],
  counts: TagBookCountResponse[],
): TagWithCount[] {
  const countById = new Map(counts.map((c) => [c.id, c.bookCount ?? 0]))
  return tags
    .map((t) => ({
      id: t.id ?? 0,
      name: t.name ?? '',
      version: t.version ?? 0,
      bookCount: countById.get(t.id) ?? 0,
    }))
    .sort((a, b) => a.name.localeCompare(b.name, 'ja'))
}
