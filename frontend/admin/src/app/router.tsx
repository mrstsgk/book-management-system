import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AdminGuard } from '@/app/AdminGuard'
import { AdminBookEditRoute } from '@/app/routes/AdminBookEditRoute'
import { AdminBookNewRoute } from '@/app/routes/AdminBookNewRoute'
import { AdminLoginRoute } from '@/app/routes/AdminLoginRoute'
import { AdminBooksListRoute } from '@/app/routes/AdminBooksListRoute'
import { AdminTagsRoute } from '@/app/routes/AdminTagsRoute'
import { NotFoundPage } from '@/components/pages/NotFoundPage'

// テストで MemoryRouter に載せられるよう、Router と分けている
export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<AdminLoginRoute />} />
      <Route element={<AdminGuard />}>
        <Route path="/" element={<Navigate to="/books" replace />} />
        <Route path="/books" element={<AdminBooksListRoute />} />
        <Route path="/books/new" element={<AdminBookNewRoute />} />
        <Route path="/books/:id/edit" element={<AdminBookEditRoute />} />
        <Route path="/tags" element={<AdminTagsRoute />} />
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
