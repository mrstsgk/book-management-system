import type { ReactNode } from 'react'

type FormFieldProps = {
  id: string
  label: string
  required?: boolean
  hint?: string
  error?: string
  count?: { current: number; max: number }
  children: ReactNode
}

// 入力に付ける aria-describedby。無い要素の id は読み上げで無視されるので、常に3つとも並べてよい
export function describedBy(id: string): string {
  return `${id}-hint ${id}-error ${id}-count`
}

export function FormField({
  id,
  label,
  required = false,
  hint,
  error,
  count,
  children,
}: FormFieldProps) {
  const over = count !== undefined && count.current > count.max
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-bold">
        {label}{' '}
        {required ? (
          <span className="text-brand-700">（必須）</span>
        ) : (
          <span className="font-normal text-ink-600">（任意）</span>
        )}
      </label>
      {children}
      {hint && (
        <p id={`${id}-hint`} className="text-[13px] text-ink-600">
          {hint}
        </p>
      )}
      {(error || count) && (
        <div className="flex justify-between gap-4 text-[13px]">
          <span id={`${id}-error`} className="text-brand-800">
            {error}
          </span>
          {count && (
            <span
              id={`${id}-count`}
              className={
                over ? 'font-bold text-brand-800' : 'text-ink-600'
              }
            >
              {count.current} / {count.max}
            </span>
          )}
        </div>
      )}
    </div>
  )
}
