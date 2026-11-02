import { Outlet } from 'react-router-dom'
import { Footer } from '@/components/layouts/Footer'
import { Header } from '@/components/layouts/Header'

export function Layout() {
  return (
    <div className="flex min-h-full flex-col bg-ink-100">
      <Header />
      <main className="flex-1">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
