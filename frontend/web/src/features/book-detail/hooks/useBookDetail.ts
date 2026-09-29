import { useGetApiBooksId } from '@/api/generated/api'
import { parseBookId } from '../utils/parseBookId'

export function useBookDetail(rawId: string | undefined) {
  const id = parseBookId(rawId)
  // ID が不正なら API を呼ばない（無効化中の 0 は使われない）
  const query = useGetApiBooksId(id ?? 0, {
    query: { enabled: id !== undefined },
  })
  const notFound = id === undefined || query.error?.status === 404

  return {
    book: query.data,
    notFound,
    isLoading: !notFound && query.isPending,
    isError: !notFound && query.isError,
    retry: () => void query.refetch(),
  }
}
