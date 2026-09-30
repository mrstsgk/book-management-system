import type { BookFormErrors, BookFormValues } from '../types'

const SUMMARY_MAX = 100
const COMMENT_MAX = 5000
const TITLE_OVERRIDE_MAX = 255
const TAG_IDS_MAX = 10

// サーバーの上限（バリデーション）と同じ値をフロントでも確かめる（送信前に気づけるように）
export function validateBookForm(values: BookFormValues): BookFormErrors {
  const errors: BookFormErrors = {}

  if (values.summary.trim() === '') {
    errors.summary = '一言まとめを入力してください'
  } else if (values.summary.length > SUMMARY_MAX) {
    errors.summary = `${SUMMARY_MAX}文字以内にしてください`
  }

  if (values.comment.trim() === '') {
    errors.comment = '感想を入力してください'
  } else if (values.comment.length > COMMENT_MAX) {
    errors.comment = `${COMMENT_MAX}文字以内にしてください`
  }

  if (values.titleOverride.length > TITLE_OVERRIDE_MAX) {
    errors.titleOverride = `${TITLE_OVERRIDE_MAX}文字以内にしてください`
  }

  if (values.rating < 1 || values.rating > 5) {
    errors.rating = '評価を選んでください'
  }

  if (values.tagIds.length > TAG_IDS_MAX) {
    errors.tagIds = `分野タグは${TAG_IDS_MAX}個までにしてください`
  }

  return errors
}
