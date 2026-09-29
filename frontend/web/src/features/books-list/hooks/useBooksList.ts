import { useSearchParams } from 'react-router-dom'
import { useGetApiBooksInfinite, useGetApiTags } from '@/api/generated/api'
import { nextOffset, PAGE_SIZE } from '../utils/paging'
import { normalizeQuery, parseTagParam } from '../utils/searchParams'

// 検索キーワードと分野タグは URL のクエリを正にする（分野別からのリンク・戻る・再読み込みで条件を保つため）
export function useBooksList() {
  const [searchParams, setSearchParams] = useSearchParams()
  const q = normalizeQuery(searchParams.get('q') ?? '')
  const tagId = parseTagParam(searchParams.get('tag'))

  const books = useGetApiBooksInfinite(
    { limit: PAGE_SIZE, q, tagId },
    { query: { initialPageParam: 0, getNextPageParam: nextOffset } },
  )
  const tags = useGetApiTags()

  const update = (next: { q?: string; tag?: number }) => {
    const params = new URLSearchParams()
    if (next.q) params.set('q', next.q)
    if (next.tag) params.set('tag', String(next.tag))
    setSearchParams(params)
  }

  return {
    query: q,
    tagId,
    hasFilter: q !== undefined || tagId !== undefined,
    books: books.data?.pages.flatMap((p) => p.items ?? []) ?? [],
    total: books.data?.pages[0]?.total ?? 0,
    isLoading: books.isPending,
    isError: books.isError,
    retry: () => void books.refetch(),
    hasMore: books.hasNextPage,
    isLoadingMore: books.isFetchingNextPage,
    loadMore: () => void books.fetchNextPage(),
    // タグ一覧が取れなくても本の一覧は出す（タグのボタンだけ出さない）
    tags: tags.data?.items,
    setQuery: (value: string) =>
      update({ q: normalizeQuery(value), tag: tagId }),
    setTag: (tag: number | undefined) => update({ q, tag }),
    clear: () => update({}),
  }
}
