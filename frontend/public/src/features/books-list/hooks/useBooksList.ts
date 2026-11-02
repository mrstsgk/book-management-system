import { useSearchParams } from 'react-router-dom'
import { useSearchBooksInfinite, useSearchTags } from '@/api/generated/api'
import { nextOffset, PAGE_SIZE } from '@/features/books-list/utils/paging'
import {
  isQueryRejected,
  isQueryTooLong,
  normalizeQuery,
  parseTagParam,
} from '@/features/books-list/utils/searchParams'

// 検索キーワードと分野タグは URL のクエリを正にする（分野別からのリンク・戻る・再読み込みで条件を保つため）
/** 本の一覧を取得し、検索・絞り込み・追加読み込みの状態を提供する。 */
export function useBooksList() {
  const [searchParams, setSearchParams] = useSearchParams()
  const q = normalizeQuery(searchParams.get('q') ?? '')
  const tagId = parseTagParam(searchParams.get('tag'))
  // 上限を超えるキーワードは API が 400 にするので、読みに行かずに入力欄で伝える
  const tooLong = isQueryTooLong(q)

  const books = useSearchBooksInfinite(
    { limit: PAGE_SIZE, q, tagId },
    {
      query: {
        initialPageParam: 0,
        getNextPageParam: nextOffset,
        enabled: !tooLong,
      },
    },
  )
  // 数え方の食い違いなどで API が q の 400 を返したときも、エラーの画面ではなく入力欄で伝える
  const queryTooLong = tooLong || isQueryRejected(books.error)
  const tags = useSearchTags()

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
    // queryTooLong のとき画面は読み込み中・エラーより先にこれを見る（読みに行かないので isPending は true のまま）
    queryTooLong,
    isLoading: books.isPending,
    isError: books.isError,
    // タグ一覧もエラーだったときに残ったまま復元しないよう、再試行では両方を取り直す
    retry: () => {
      void books.refetch()
      void tags.refetch()
    },
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
