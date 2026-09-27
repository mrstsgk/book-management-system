import { Link } from 'react-router-dom'

export function Header() {
  return (
    <header className="sticky top-0 z-40 bg-white text-ink-900 shadow-soft">
      <div className="container-wide flex h-14 items-center md:h-16">
        <Link to="/" className="text-lg font-bold hover:text-brand-600">
          書籍管理システム
        </Link>
      </div>
    </header>
  )
}
