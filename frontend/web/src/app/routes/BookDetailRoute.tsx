import { Link, useParams } from 'react-router-dom'
import { EmptyState } from '@/components/states/EmptyState'
import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { BookDetailView } from '@/features/book-detail/components/BookDetailView'
import { useBookDetail } from '@/features/book-detail/hooks/useBookDetail'

export function BookDetailRoute() {
  const { id } = useParams()
  const { book, notFound, isLoading, isError, retry } = useBookDetail(id)

  return (
    <section className="container-wide py-8 md:py-10">
      {notFound ? (
        <EmptyState
          title="この本は見つかりませんでした"
          description="削除されたか、URL が間違っている可能性があります。"
          action={
            <Link to="/" className="font-bold text-brand-700 hover:underline">
              一覧へ戻る
            </Link>
          }
        />
      ) : isLoading ? (
        <LoadingState />
      ) : isError || !book ? (
        <ErrorState
          title="本を読み込めませんでした"
          description="時間をおいてもう一度お試しください。"
          onRetry={retry}
        />
      ) : (
        <BookDetailView book={book} />
      )}
    </section>
  )
}
