type ErrorStateProps = {
  title: string
  description?: string
  onRetry: () => void
}

export function ErrorState({ title, description, onRetry }: ErrorStateProps) {
  return (
    <div
      role="alert"
      className="card flex flex-col items-center gap-4 border border-brand-700 px-6 py-10 text-center md:px-10"
    >
      <p className="text-lg font-bold">{title}</p>
      {description && <p className="text-sm text-ink-600">{description}</p>}
      <button
        type="button"
        onClick={onRetry}
        className="h-12 rounded-lg bg-brand-700 px-8 font-bold text-white hover:bg-brand-800"
      >
        再試行
      </button>
    </div>
  )
}
