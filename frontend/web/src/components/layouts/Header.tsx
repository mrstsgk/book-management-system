import { Link, NavLink } from 'react-router-dom'

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `flex h-full items-center border-b-2 ${
    isActive
      ? 'border-brand-500 font-bold text-ink-900'
      : 'border-transparent text-ink-600 hover:text-ink-900'
  }`

export function Header() {
  return (
    <header className="sticky top-0 z-40 bg-white text-ink-900 shadow-soft">
      <div className="container-wide flex h-14 items-center justify-between gap-10 md:h-16 md:justify-start">
        <Link to="/" className="text-lg font-bold hover:text-brand-700">
          読んだ本
        </Link>
        <nav aria-label="メイン" className="flex h-full gap-5 text-sm md:gap-7 md:text-[15px]">
          {/* end: 詳細（/books/:id）を開いているときに「一覧」を今いる画面にしない */}
          <NavLink to="/" end className={navLinkClass}>
            一覧
          </NavLink>
          <NavLink to="/tags" className={navLinkClass}>
            分野別
          </NavLink>
        </nav>
      </div>
    </header>
  )
}
