import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { AdminGuard } from '@/app/AdminGuard'
import { AdminBookEditRoute } from '@/app/routes/AdminBookEditRoute'
import { AdminBookNewRoute } from '@/app/routes/AdminBookNewRoute'
import { AdminLoginRoute } from '@/app/routes/AdminLoginRoute'
import { AdminBooksListRoute } from '@/app/routes/AdminBooksListRoute'
import { AdminTagsRoute } from '@/app/routes/AdminTagsRoute'
import { BookDetailRoute } from '@/app/routes/BookDetailRoute'
import { BooksListRoute } from '@/app/routes/BooksListRoute'
import { TagCountsRoute } from '@/app/routes/TagCountsRoute'
import { Layout } from '@/components/layouts/Layout'
import { NotFoundPage } from '@/components/pages/NotFoundPage'

// テストで MemoryRouter に載せられるよう、Router と分けている
export function AppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<BooksListRoute />} />
        <Route path="/books/:id" element={<BookDetailRoute />} />
        <Route path="/tags" element={<TagCountsRoute />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
      <Route path="/admin/login" element={<AdminLoginRoute />} />
      <Route path="/admin" element={<AdminGuard />}>
        <Route index element={<AdminBooksListRoute />} />
        <Route path="books/new" element={<AdminBookNewRoute />} />
        <Route path="books/:id/edit" element={<AdminBookEditRoute />} />
        <Route path="tags" element={<AdminTagsRoute />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
