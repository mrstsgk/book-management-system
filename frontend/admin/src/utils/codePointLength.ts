/**
 * 文字列の長さを Unicode のコードポイントで数える。
 * BE（Go の utf8.RuneCountInString）・OpenAPI の maxLength・生成 zod の .max() と同じ数え方。
 * 入力欄の文字数の表示に UTF-16 単位の String.length（絵文字を 2 と数える）を使うと、上限とずれるため。
 */
export function codePointLength(value: string): number {
  return Array.from(value).length
}
