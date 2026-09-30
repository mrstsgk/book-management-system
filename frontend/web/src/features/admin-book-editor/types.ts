// 本の登録・編集フォームの値。登録は空から、編集は既存の本の値から始める
export type BookFormValues = {
  titleOverride: string
  summary: string
  comment: string
  rating: number
  tagIds: number[]
}

export type BookFormErrors = Partial<Record<keyof BookFormValues, string>>

export const emptyBookFormValues: BookFormValues = {
  titleOverride: '',
  summary: '',
  comment: '',
  rating: 0,
  tagIds: [],
}
