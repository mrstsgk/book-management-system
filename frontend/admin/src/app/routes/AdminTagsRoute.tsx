import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { TagsView } from '@/features/admin-tags/components/TagsView'
import { useAdminTags } from '@/features/admin-tags/hooks/useAdminTags'

export function AdminTagsRoute() {
  const { items, isPending, isError, refetch, addTag, renameTag, deleteTag } =
    useAdminTags()

  return (
    <section className="container-wide flex flex-col gap-6 py-10">
      <h1 className="text-[28px] font-extrabold">分野タグ</h1>
      {isPending ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState
          title="タグを読み込めませんでした"
          description="時間をおいてもう一度お試しください。"
          onRetry={refetch}
        />
      ) : (
        <TagsView
          items={items}
          onAdd={addTag}
          onRename={renameTag}
          onDelete={deleteTag}
        />
      )}
    </section>
  )
}
