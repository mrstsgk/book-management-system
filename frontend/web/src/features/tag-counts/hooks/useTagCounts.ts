import { useGetApiTagsCounts } from '@/api/generated/api'

export function useTagCounts() {
  const { data, isPending, isError, refetch } = useGetApiTagsCounts()

  return {
    items: data?.items ?? [],
    isPending,
    isError,
    refetch: () => void refetch(),
  }
}
