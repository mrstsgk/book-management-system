import { Link, Outlet, useLocation } from 'react-router-dom'

const navClass = (current: boolean) =>
  `flex h-full items-center border-b-2 ${
    current
      ? 'border-brand-400 font-bold text-white'
      : 'border-transparent text-ink-300 hover:text-white'
  }`

interface AdminLayoutProps {
  onLogout: () => void
  // ログアウトに失敗したとき true。サーバー側のセッションが残っているので画面は抜けず、再試行を促す
  logoutFailed?: boolean
}

// 本番は VITE_PUBLIC_SITE_URL で公開画面の URL を渡す。ローカルの開発サーバーの既定値は :3000
const publicSiteUrl =
  import.meta.env.VITE_PUBLIC_SITE_URL ?? 'http://localhost:3000'

export function AdminLayout({
  onLogout,
  logoutFailed = false,
}: AdminLayoutProps) {
  const { pathname } = useLocation()
  // 「本」は一覧・登録・編集のどれでも今いる画面にする（NavLink の前方一致だと /tags まで当たる）
  const onBooks = pathname === '/books' || pathname.startsWith('/books')
  const onTags = pathname.startsWith('/tags')

  return (
    <div className="flex min-h-full flex-col bg-ink-100">
      <header className="sticky top-0 z-40 bg-ink-900 text-white">
        <div className="container-wide flex h-14 items-center gap-6 md:h-16 md:gap-10">
          <Link
            to="/books"
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
              to="/books"
              aria-current={onBooks ? 'page' : undefined}
              className={navClass(onBooks)}
            >
              本
            </Link>
            <Link
              to="/tags"
              aria-current={onTags ? 'page' : undefined}
              className={navClass(onTags)}
            >
              タグ
            </Link>
          </nav>
          {/* 公開画面は別のアプリ（frontend/public）なので、ルーターの Link ではなく URL で開く */}
          <a
            href={publicSiteUrl}
            className="text-sm text-ink-300 hover:text-white"
          >
            公開画面を見る
          </a>
          <button
            type="button"
            onClick={onLogout}
            className="text-sm text-ink-300 hover:text-white"
          >
            ログアウト
          </button>
        </div>
      </header>
      {logoutFailed && (
        <div
          role="alert"
          className="border-b border-brand-700 bg-brand-50 px-4 py-3 text-center text-sm text-brand-800"
        >
          ログアウトできませんでした。ログイン状態は残っています。もう一度「ログアウト」を押してください。
        </div>
      )}
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  )
}
