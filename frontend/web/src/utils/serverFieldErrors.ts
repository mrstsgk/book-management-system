import type { FieldErrorDetail } from '@/api/mutator'

const GENERIC_MESSAGE = '入力を見直してください'

// field ごとの rule → 日本語の文言。フロントの validateBookForm と表記を揃える
const MESSAGES: Record<string, Record<string, string>> = {
  isbn: {
    required: 'ISBNを入力してください',
    max: '17文字以内にしてください',
  },
  summary: {
    required: '一言まとめを入力してください',
    max: '100文字以内にしてください',
  },
  comment: {
    required: '感想を入力してください',
    max: '5000文字以内にしてください',
  },
  rating: {
    required: '評価を選んでください',
    min: '評価を選んでください',
    max: '評価を選んでください',
  },
  tagIds: {
    max: '分野タグは10個までにしてください',
  },
  titleOverride: {
    max: '255文字以内にしてください',
  },
}

// サーバーの errors[{field, rule}] を、画面に出す文言（field ごと）に変える
export function toFieldMessages(
  errors: FieldErrorDetail[],
): Record<string, string> {
  const out: Record<string, string> = {}
  for (const e of errors) {
    out[e.field] = MESSAGES[e.field]?.[e.rule] ?? GENERIC_MESSAGE
  }
  return out
}
