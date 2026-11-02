import {
  registerBookBodyCommentMax,
  registerBookBodyRatingMax,
  registerBookBodySummaryMax,
  registerBookBodyTagIdsMax,
  registerBookBodyTitleOverrideMax,
} from '@/api/generated/zod/api'
import type {
  BookFormErrors,
  BookFormValues,
} from '@/features/admin-book-editor/types'
import { codePointLength } from '@/utils/codePointLength'

// 上限の値は OpenAPI から生成した zod の定数を使い、FE で書かない（FE・BE・契約で値がずれないように）。
// 長さは BE と同じコードポイントで数える（String.length は絵文字を 2 と数えるため）。文言はこの画面の言い回しに
// 合わせて手で持つので、zod のスキーマ全体ではなく定数だけを使う
export function validateBookForm(values: BookFormValues): BookFormErrors {
  const errors: BookFormErrors = {}

  if (values.summary.trim() === '') {
    errors.summary = '一言まとめを入力してください'
  } else if (codePointLength(values.summary) > registerBookBodySummaryMax) {
    errors.summary = `${registerBookBodySummaryMax}文字以内にしてください`
  }

  if (values.comment.trim() === '') {
    errors.comment = '感想を入力してください'
  } else if (codePointLength(values.comment) > registerBookBodyCommentMax) {
    errors.comment = `${registerBookBodyCommentMax}文字以内にしてください`
  }

  if (
    codePointLength(values.titleOverride) > registerBookBodyTitleOverrideMax
  ) {
    errors.titleOverride = `${registerBookBodyTitleOverrideMax}文字以内にしてください`
  }

  if (values.rating < 1 || values.rating > registerBookBodyRatingMax) {
    errors.rating = '評価を選んでください'
  }

  if (values.tagIds.length > registerBookBodyTagIdsMax) {
    errors.tagIds = `分野タグは${registerBookBodyTagIdsMax}個までにしてください`
  }

  return errors
}
