import { useLocation } from 'react-router-dom'
import { AdminBooksListView } from '@/features/admin-books-list/components/AdminBooksListView'
import { useAdminBooksList } from '@/features/admin-books-list/hooks/useAdminBooksList'

export function AdminBooksListRoute() {
  const location = useLocation()
  const list = useAdminBooksList()
  return (
    <AdminBooksListView
      books={list.books}
      total={list.total}
      // 登録・保存・削除の後にこの画面へ渡す完了のお知らせ（design.md「本の登録・編集」）。
      // ponytail: 戻る操作で history state が残り、古い notice が再表示され得るが、
      // 個人用の管理画面での軽微な見た目のずれなので許容する
      notice={(location.state as { notice?: string } | null)?.notice}
      isLoading={list.isLoading}
      isError={list.isError}
      onRetry={list.retry}
      hasMore={list.hasMore}
      isLoadingMore={list.isLoadingMore}
      onLoadMore={list.loadMore}
    />
  )
}
