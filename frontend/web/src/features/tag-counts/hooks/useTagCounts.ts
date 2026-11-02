import { useCountBooks } from '@/api/generated/api'

export function useTagCounts() {
  const { data, isPending, isError, refetch } = useCountBooks()

  return {
    items: data?.items ?? [],
    isPending,
    isError,
    refetch: () => void refetch(),
  }
}
