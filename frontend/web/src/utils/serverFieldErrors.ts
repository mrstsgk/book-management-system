import type { FieldErrorDetail } from '@/api/mutator'

const GENERIC_MESSAGE = '入力を見直してください'

// field ごとの rule → 日本語の文言。rule は validate タグのキーワード（required / max …）、
// VO 由来の format、相関ルールの id（docs/design/admin-screens.md の CR-xx と MSG-Axx）。
// 本の登録・編集とタグの両方で使うため、feature ではなく共有に置く
const MESSAGES: Record<string, Record<string, string>> = {
  isbn: {
    required: 'ISBNを入力してください',
    max: '17文字以内にしてください',
    format: 'ISBN は 13 桁か 10 桁で入力してください。',
    isbnDuplicated: 'この ISBN の本はすでに登録されています。',
    isbnNotInCatalog:
      'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
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
    tagIdsNotFound:
      '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
  },
  titleOverride: {
    max: '255文字以内にしてください',
  },
  name: {
    required: 'タグの名前を入力してください',
    max: '30文字以内にしてください',
    tagNameDuplicated:
      'この名前の分野タグはすでにあります。別の名前にしてください。',
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
