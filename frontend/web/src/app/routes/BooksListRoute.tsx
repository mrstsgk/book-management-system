import { BooksListView } from '@/features/books-list/components/BooksListView'
import { useBooksList } from '@/features/books-list/hooks/useBooksList'

export function BooksListRoute() {
  const list = useBooksList()
  return (
    <BooksListView
      books={list.books}
      total={list.total}
      query={list.query}
      tagId={list.tagId}
      tags={list.tags}
      hasFilter={list.hasFilter}
      isLoading={list.isLoading}
      isError={list.isError}
      onRetry={list.retry}
      hasMore={list.hasMore}
      isLoadingMore={list.isLoadingMore}
      onLoadMore={list.loadMore}
      onSearch={list.setQuery}
      onSelectTag={list.setTag}
      onClear={list.clear}
    />
  )
}
