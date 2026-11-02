import type { ApiError } from '@/api/mutator'
import { codePointLength } from '@/utils/codePointLength'

// URL の tag は正の整数だけを受け付ける。それ以外は API の 400 を画面に出さないよう絞り込みなしにする
export function parseTagParam(value: string | null): number | undefined {
  if (value === null || !/^[1-9]\d*$/.test(value)) return undefined
  // 桁の多い数字は丸められて別のIDになるため、安全な整数に収まらなければ絞り込みなしにする
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : undefined
}

// 上限は API の q（Domain の検索キーワード）と同じ。String.length は UTF-16 単位で絵文字を 2 と数えるため使わない
// 契約（openapi/）に q の maxLength は書けない（前後の空白を除いてから数えるため）。上限は BE の VO と同じ値を持つ
const MAX_QUERY_LENGTH = 100

/** 検索キーワードが上限（前後の空白を除いてコードポイントで数える）を超えるかを判定する。 */
export function isQueryTooLong(value: string | undefined): boolean {
  return (
    value !== undefined && codePointLength(value.trim()) > MAX_QUERY_LENGTH
  )
}

/** API が検索キーワード（q）を入力の誤りとして 400 で返したかを判定する。 */
export function isQueryRejected(error: ApiError | null): boolean {
  return error?.status === 400 && error.fieldErrors.some((e) => e.field === 'q')
}

/** 検索キーワードの前後の空白を除き、空ならキーワードなし（undefined）にする。 */
export function normalizeQuery(value: string): string | undefined {
  const trimmed = value.trim()
  return trimmed === '' ? undefined : trimmed
}
