import { Link } from 'react-router-dom'
import type { BookListItemResponse } from '@/api/generated/api.schemas'
import { EmptyState } from '@/components/states/EmptyState'
import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { RatingStars } from '@/components/ui/RatingStars'

type AdminBooksListViewProps = {
  books: BookListItemResponse[]
  total: number
  notice?: string
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  hasMore: boolean
  isLoadingMore: boolean
  onLoadMore: () => void
}

const registerLink = (
  <Link
    to="/admin/books/new"
    className="inline-flex h-12 items-center rounded-lg bg-brand-700 px-6 font-bold text-white hover:bg-brand-800"
  >
    ＋ 本を登録
  </Link>
)

export function AdminBooksListView(props: AdminBooksListViewProps) {
  const { total, notice, isLoading, isError, onRetry } = props

  return (
    <section className="container-wide flex flex-col gap-5 py-8 md:py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold md:text-[28px]">本</h1>
        {registerLink}
      </div>
      {notice && (
        <p role="status" className="card px-5 py-3.5 text-sm">
          {notice}
        </p>
      )}
      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState
          title="本の一覧を読み込めませんでした"
          description="サーバーに接続できません。時間をおいてもう一度お試しください。"
          onRetry={onRetry}
        />
      ) : total === 0 ? (
        <EmptyState title="まだ本が登録されていません" />
      ) : (
        <BooksTable {...props} />
      )}
    </section>
  )
}

function BooksTable({
  books,
  total,
  hasMore,
  isLoadingMore,
  onLoadMore,
}: AdminBooksListViewProps) {
  return (
    <div className="flex flex-col items-center gap-4">
      <table className="card w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-ink-200 bg-ink-50 text-left text-ink-600">
            <th scope="col" className="px-5 py-3.5 font-bold">
              書名
            </th>
            <th scope="col" className="w-[120px] px-5 py-3.5 font-bold">
              評価
            </th>
            <th scope="col" className="w-[220px] px-5 py-3.5 font-bold">
              分野タグ
            </th>
            <th scope="col" className="w-[140px] px-5 py-3.5 font-bold">
              一言まとめ
            </th>
            <th scope="col" className="w-[100px] px-5 py-3.5">
              <span className="sr-only">操作</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {books.map((book) => (
            <tr key={book.id} className="border-t border-ink-200">
              <td className="px-5 py-4 font-bold">{book.title}</td>
              <td className="px-5 py-4 text-brand-700">
                <RatingStars rating={book.rating ?? 0} />
              </td>
              <td className="px-5 py-4 text-ink-600">
                {book.tags && book.tags.length > 0 ? book.tags.join('・') : '—'}
              </td>
              <td className="px-5 py-4 text-ink-600">
                {book.summary && book.summary !== '（準備中）'
                  ? '記入済み'
                  : '未記入'}
              </td>
              <td className="px-5 py-4 text-right">
                <Link to={`/admin/books/${book.id}/edit`} className="font-bold">
                  編集
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="text-[13px] text-ink-600">
        {total}冊中 {books.length}冊を表示
      </p>
      {hasMore && (
        <>
          {isLoadingMore && (
            <p role="status" className="text-sm text-ink-700">
              続きを読み込み中…
            </p>
          )}
          <button
            type="button"
            onClick={onLoadMore}
            disabled={isLoadingMore}
            className="h-12 rounded-lg border border-brand-700 px-10 font-bold text-brand-700 disabled:cursor-not-allowed disabled:border-ink-400 disabled:text-ink-500"
          >
            もっと見る
          </button>
        </>
      )}
    </div>
  )
}
