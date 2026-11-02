import type { FieldErrorDetail } from '@/api/mutator'

const GENERIC_MESSAGE = '入力を見直してください'

// API が返すコードから画面の文言を引く対応表（frontend/architecture.md §2.2）。キーは 2 種類:
// - `<field>.<rule>`: 400 の errors[{field, rule}]。rule は OpenAPI のキーワード（required / maxLength …）、
//   VO 由来の format、相関ルールの id（x-correlation-rules。docs/design/admin-screens.md の CR-xx）
// - `<code>`: 業務ルールの違反（x-error-codes）。{キー} には params の値を差し込む
// 本の登録・編集・タグ・ログインで使うため、feature ではなく共有に置く
export const ERROR_MESSAGES: Record<string, string> = {
  'isbn.required': 'ISBNを入力してください',
  'isbn.maxLength': '17文字以内にしてください',
  'isbn.format': 'ISBN は 13 桁か 10 桁で入力してください。',
  'isbn.isbnDuplicated': 'この ISBN の本はすでに登録されています。',
  'isbn.isbnNotInCatalog':
    'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
  'summary.required': '一言まとめを入力してください',
  'summary.maxLength': '100文字以内にしてください',
  'comment.required': '感想を入力してください',
  'comment.maxLength': '5000文字以内にしてください',
  'rating.required': '評価を選んでください',
  'rating.minimum': '評価を選んでください',
  'rating.maximum': '評価を選んでください',
  'tagIds.maxItems': '分野タグは10個までにしてください',
  'tagIds.tagIdsNotFound':
    '選んだ分野タグのうち、すでに削除されたものがあります。タグを選び直してください。',
  'titleOverride.maxLength': '255文字以内にしてください',
  'name.required': 'タグの名前を入力してください',
  'name.maxLength': '30文字以内にしてください',
  'name.tagNameDuplicated':
    'この名前の分野タグはすでにあります。別の名前にしてください。',
  bookNotFound: 'この本は見つかりませんでした',
  tagNotFound: 'この分野タグはすでに削除されています。',
  catalogEntryNotFound:
    'この ISBN の本は外部カタログに見つかりませんでした。ISBN を確かめてください。',
  versionConflict:
    '他で先に更新されています。最新の内容を読み込んでから、もう一度保存してください。',
  unauthenticated: 'ログインしてください',
  loginFailed: 'IDかパスワードが違います',
  tooManyLoginAttempts:
    'ログインに続けて失敗したため、{retryAfterSeconds}秒ほど待ってからやり直してください',
  isbnDuplicated: 'この ISBN の本はすでに登録されています。',
  tagNameDuplicated:
    'この名前の分野タグはすでにあります。別の名前にしてください。',
}

/** コードに対応する文言を、params を差し込んで返す。対応が無い・差し込む値が足りなければ undefined。 */
export function getErrorMessage(
  key: string,
  params: Record<string, unknown> = {},
): string | undefined {
  const template = ERROR_MESSAGES[key]
  if (template === undefined) return undefined
  // 値の欠けた {キー} を画面に出さないよう、呼び出し側の既定の文言に任せる
  const names = [...template.matchAll(/\{(\w+)\}/g)].map((m) => m[1])
  if (names.some((name) => !(name in params))) return undefined
  return template.replace(/\{(\w+)\}/g, (_, name: string) =>
    String(params[name]),
  )
}

/** サーバーの errors[{field, rule}] を、画面に出す文言（field ごと）に変える。 */
export function toFieldMessages(
  errors: FieldErrorDetail[],
): Record<string, string> {
  const out: Record<string, string> = {}
  for (const e of errors) {
    out[e.field] =
      getErrorMessage(`${e.field}.${e.rule}`, e.params) ?? GENERIC_MESSAGE
  }
  return out
}
