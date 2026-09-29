type LoadingStateProps = {
  label?: string
}

export function LoadingState({ label = '読み込み中…' }: LoadingStateProps) {
  return (
    <p role="status" className="py-10 text-center text-[15px] text-ink-700">
      {label}
    </p>
  )
}
