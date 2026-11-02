import { EmptyState } from '@/components/states/EmptyState'
import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { TagCountsView } from '@/features/tag-counts/components/TagCountsView'
import { useTagCounts } from '@/features/tag-counts/hooks/useTagCounts'

export function TagCountsRoute() {
  const { items, isPending, isError, refetch } = useTagCounts()

  return (
    <section className="container-wide flex flex-col gap-6 py-10">
      <div className="flex flex-col gap-1.5">
        <h1 className="text-[28px] font-extrabold">分野別</h1>
        <p className="text-sm text-ink-600">
          分野を選ぶと、その分野の本の一覧へ移ります。
        </p>
      </div>
      {isPending ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState
          title="分野別の冊数を読み込めませんでした"
          description="時間をおいてもう一度お試しください。"
          onRetry={refetch}
        />
      ) : items.length === 0 ? (
        <EmptyState title="分野タグの付いた本はまだありません" />
      ) : (
        <>
          <TagCountsView items={items} />
          <p className="text-[13px] text-ink-600">
            冊数の多い順。本が付いていない分野は表示しません。
          </p>
        </>
      )}
    </section>
  )
}
