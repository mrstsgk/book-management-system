import { BrowserRouter, Route, Routes } from 'react-router-dom'
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
