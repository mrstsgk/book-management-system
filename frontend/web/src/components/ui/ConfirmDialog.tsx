import { useEffect, useId, useRef, type KeyboardEvent } from 'react'

type ConfirmDialogProps = {
  open: boolean
  title: string
  description: string
  confirmLabel: string
  onConfirm: () => void
  onCancel: () => void
  busy?: boolean
}

// ネイティブの <dialog> を showModal() で開き、背景の操作とフォーカスの閉じ込めはブラウザに任せる
export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  onConfirm,
  onCancel,
  busy = false,
}: ConfirmDialogProps) {
  const ref = useRef<HTMLDialogElement>(null)
  const cancelRef = useRef<HTMLButtonElement>(null)
  const titleId = useId()
  const descriptionId = useId()

  useEffect(() => {
    const dialog = ref.current
    if (!dialog) return
    if (open && !dialog.open) {
      dialog.showModal()
      // 取り消せない操作の確認なので、既定のフォーカスは確定ではなくキャンセルに置く
      cancelRef.current?.focus()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open])

  // Esc はブラウザが dialog を勝手に閉じる前に止め、閉じるかどうかを呼ぶ側（open）に任せる
  const onKeyDown = (e: KeyboardEvent<HTMLDialogElement>) => {
    if (e.key !== 'Escape') return
    e.preventDefault()
    onCancel()
  }

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
      onKeyDown={onKeyDown}
      onCancel={(e) => {
        e.preventDefault()
        onCancel()
      }}
      className="w-[min(480px,calc(100vw-32px))] rounded-xl p-8 shadow-soft backdrop:bg-ink-900/45"
    >
      <div className="flex flex-col gap-4">
        <h2 id={titleId} className="text-xl font-extrabold">
          {title}
        </h2>
        <p id={descriptionId} className="leading-relaxed">
          {description}
        </p>
        <div className="flex justify-end gap-3">
          <button
            ref={cancelRef}
            type="button"
            onClick={onCancel}
            className="h-11 rounded-lg border border-ink-400 bg-white px-5"
          >
            キャンセル
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={busy}
            className="h-11 rounded-lg bg-brand-700 px-5 font-bold text-white hover:bg-brand-800 disabled:cursor-not-allowed disabled:bg-ink-400"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </dialog>
  )
}
