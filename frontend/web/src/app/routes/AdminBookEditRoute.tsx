import { useParams } from 'react-router-dom'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { EmptyState } from '@/components/states/EmptyState'
import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { BookForm } from '@/features/admin-book-editor/components/BookForm'
import { useBookEditor } from '@/features/admin-book-editor/hooks/useBookEditor'

export function AdminBookEditRoute() {
  const { id } = useParams()
  const editor = useBookEditor(id)

  if (editor.notFound) {
    return (
      <section className="container-wide py-8 md:py-10">
        <EmptyState
          title="この本は見つかりませんでした"
          description="削除されたか、URL が間違っている可能性があります。"
        />
      </section>
    )
  }
  if (editor.isLoading) {
    return (
      <section className="container-wide py-8 md:py-10">
        <LoadingState />
      </section>
    )
  }
  if (editor.isError || !editor.book) {
    return (
      <section className="container-wide py-8 md:py-10">
        <ErrorState
          title="本を読み込めませんでした"
          description="時間をおいてもう一度お試しください。"
          onRetry={editor.retry}
        />
      </section>
    )
  }

  return (
    <section className="container-wide flex flex-col gap-6 py-8 md:py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold">本を編集</h1>
        <button
          type="button"
          onClick={editor.openDelete}
          className="h-11 rounded-lg border border-brand-700 px-5 text-[15px] font-bold text-brand-700"
        >
          この本を削除
        </button>
      </div>

      <div className="card flex gap-5 p-6 md:p-7">
        <div className="w-24 flex-shrink-0">
          {editor.book.coverUrl ? (
            <img
              src={editor.book.coverUrl}
              alt={`${editor.book.title ?? ''}の書影`}
              className="w-24 rounded"
            />
          ) : (
            <div className="flex h-[136px] w-24 items-center justify-center rounded bg-ink-200 text-xs text-ink-600">
              書影なし
            </div>
          )}
        </div>
        <dl className="grid grid-cols-[80px_1fr] gap-y-1.5 text-sm">
          <dt className="text-ink-600">書名</dt>
          <dd className="font-bold">{editor.book.title}</dd>
          {editor.book.authors && (
            <>
              <dt className="text-ink-600">著者</dt>
              <dd>{editor.book.authors}</dd>
            </>
          )}
          <dt className="text-ink-600">ISBN</dt>
          <dd>{editor.book.isbn}</dd>
        </dl>
      </div>

      {editor.conflict && (
        <div
          role="alert"
          className="flex items-center justify-between gap-4 rounded-lg border border-brand-700 bg-brand-50 px-5 py-3.5 text-sm text-brand-800"
        >
          <span>
            この本は、ほかの画面で先に更新されていたため保存しませんでした。
          </span>
          <button
            type="button"
            onClick={() => void editor.reloadLatest()}
            className="h-10 flex-shrink-0 rounded-lg bg-brand-700 px-4 font-bold text-white"
          >
            最新を読み込む
          </button>
        </div>
      )}

      {editor.values && (
        <div className="card p-6 md:p-7">
          <BookForm
            values={editor.values}
            onChange={editor.setValues}
            errors={editor.errors}
            tags={editor.tags}
            submitLabel="保存する"
            onSubmit={editor.submit}
            submitting={editor.submitting}
          />
        </div>
      )}

      <ConfirmDialog
        open={editor.deleteOpen}
        title={`「${editor.book.title ?? ''}」を削除しますか？`}
        description="この本の一言まとめ・感想・評価を削除します。元に戻せません。"
        confirmLabel="削除する"
        onConfirm={editor.confirmDelete}
        onCancel={editor.closeDelete}
        busy={editor.deleting}
      />
    </section>
  )
}
