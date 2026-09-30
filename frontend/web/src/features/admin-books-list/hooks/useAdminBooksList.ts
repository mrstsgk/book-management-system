import { useGetApiBooksInfinite } from '@/api/generated/api'
import { nextOffset, PAGE_SIZE } from '../utils/paging'

// 管理画面の本の一覧。検索・絞り込みは無い（冊数が少ないため。要件定義 §3・設計書「やらないこと」）
export function useAdminBooksList() {
  const books = useGetApiBooksInfinite(
    { limit: PAGE_SIZE },
    { query: { initialPageParam: 0, getNextPageParam: nextOffset } },
  )

  return {
    books: books.data?.pages.flatMap((p) => p.items ?? []) ?? [],
    total: books.data?.pages[0]?.total ?? 0,
    isLoading: books.isPending,
    isError: books.isError,
    retry: () => void books.refetch(),
    hasMore: books.hasNextPage,
    isLoadingMore: books.isFetchingNextPage,
    loadMore: () => void books.fetchNextPage(),
  }
}
