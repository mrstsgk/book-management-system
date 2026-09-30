type ErrorSummaryProps = {
  errors: { id: string; message: string }[]
}

export function ErrorSummary({ errors }: ErrorSummaryProps) {
  if (errors.length === 0) return null
  return (
    <div
      role="alert"
      className="flex flex-col gap-1 rounded-lg border border-brand-700 bg-brand-50 px-5 py-3.5 text-sm text-brand-800"
    >
      <strong>{errors.length}か所の入力を直してください。</strong>
      {errors.map((e) => (
        <a key={e.id} href={`#${e.id}`} className="text-brand-800 underline">
          {e.message}
        </a>
      ))}
    </div>
  )
}
