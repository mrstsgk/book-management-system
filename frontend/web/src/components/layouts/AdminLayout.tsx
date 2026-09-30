import { Link, Outlet, useLocation } from 'react-router-dom'
import { adminToken } from '@/lib/admin-auth'

const navClass = (current: boolean) =>
  `flex h-full items-center border-b-2 ${
    current
      ? 'border-brand-400 font-bold text-white'
      : 'border-transparent text-ink-300 hover:text-white'
  }`

export function AdminLayout() {
  const { pathname } = useLocation()
  // 「本」は一覧・登録・編集のどれでも今いる画面にする（NavLink の前方一致だと /admin/tags まで当たる）
  const onBooks = pathname === '/admin' || pathname.startsWith('/admin/books')
  const onTags = pathname.startsWith('/admin/tags')

  return (
    <div className="flex min-h-full flex-col bg-ink-100">
      <header className="sticky top-0 z-40 bg-ink-900 text-white">
        <div className="container-wide flex h-14 items-center gap-6 md:h-16 md:gap-10">
          <Link
            to="/admin"
            className="flex items-center gap-2.5 text-lg font-bold"
          >
            読んだ本
            <span className="rounded bg-white px-2 py-0.5 text-xs font-bold text-ink-900">
              管理
            </span>
          </Link>
          <nav
            aria-label="管理メニュー"
            className="flex h-full flex-1 gap-5 text-sm md:gap-7 md:text-[15px]"
          >
            <Link
              to="/admin"
              aria-current={onBooks ? 'page' : undefined}
              className={navClass(onBooks)}
            >
              本
            </Link>
            <Link
              to="/admin/tags"
              aria-current={onTags ? 'page' : undefined}
              className={navClass(onTags)}
            >
              タグ
            </Link>
          </nav>
          <Link to="/" className="text-sm text-ink-300 hover:text-white">
            公開画面を見る
          </Link>
        </div>
      </header>
      {!adminToken() && (
        <div
          role="alert"
          className="border-b border-brand-700 bg-brand-50 px-4 py-3 text-center text-sm text-brand-800"
        >
          管理者トークンが設定されていません。
          <code className="mx-1">frontend/web/.env.development.local</code>に
          <code className="mx-1">VITE_ADMIN_TOKEN</code>
          を書いて開発サーバーを起動し直すまで、書き込みはできません。
        </div>
      )}
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  )
}
