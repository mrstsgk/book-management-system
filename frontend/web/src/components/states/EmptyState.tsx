import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="card flex flex-col items-center gap-4 px-6 py-10 text-center md:px-10">
      <p className="text-lg font-bold">{title}</p>
      {description && <p className="text-sm text-ink-600">{description}</p>}
      {action}
    </div>
  )
}
