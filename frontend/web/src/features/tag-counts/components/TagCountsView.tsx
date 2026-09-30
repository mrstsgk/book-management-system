import { Link } from 'react-router-dom'
import type { TagBookCountResponse } from '@/api/generated/api.schemas'
import { barRatio } from '../utils/barRatio'

type TagCountsViewProps = {
  items: TagBookCountResponse[]
}

export function TagCountsView({ items }: TagCountsViewProps) {
  const max = Math.max(0, ...items.map((item) => item.bookCount ?? 0))

  return (
    <ul className="card flex max-w-[880px] flex-col py-3">
      {items.map((item) => (
        <li key={item.id}>
          <Link
            to={`/?tag=${item.id}`}
            className="grid min-h-14 grid-cols-[minmax(0,7rem)_minmax(0,1fr)_3.5rem] items-center gap-4 px-4 py-3.5 text-ink-900 hover:bg-ink-50 md:grid-cols-[180px_minmax(0,1fr)_64px] md:gap-5 md:px-7"
          >
            <span className="truncate font-bold">{item.name}</span>
            <span
              aria-hidden="true"
              className="flex h-3 overflow-hidden rounded-pill bg-ink-100"
            >
              <span
                className="block h-3 rounded-pill bg-brand-600"
                style={{ width: `${barRatio(item.bookCount ?? 0, max)}%` }}
              />
            </span>
            <span className="text-right">{item.bookCount}冊</span>
          </Link>
        </li>
      ))}
    </ul>
  )
}
