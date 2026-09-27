import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { Layout } from '@/components/layouts/Layout'
import { ComingSoonPage } from '@/components/pages/ComingSoonPage'
import { HomePage } from '@/features/home/components/HomePage'

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="*" element={<ComingSoonPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
