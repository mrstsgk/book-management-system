import type { BookFormErrors, BookFormValues } from '../types'
import { validateBookForm } from './validateBookForm'

// 送信時に検出したエラー（localErrors）のうち、今の入力値でまだ誤りが残っている項目だけを残す。
// 直した項目は次の送信を待たずに消え、直していない項目はそのまま表示され続ける
export function visibleLocalErrors(
  localErrors: BookFormErrors,
  values: BookFormValues,
): BookFormErrors {
  const liveErrors = validateBookForm(values)
  return Object.fromEntries(
    (Object.keys(localErrors) as (keyof BookFormErrors)[])
      .filter((field) => field in liveErrors)
      .map((field) => [field, liveErrors[field]]),
  )
}
