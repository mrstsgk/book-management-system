// URL の tag は正の整数だけを受け付ける。それ以外は API の 400 を画面に出さないよう絞り込みなしにする
export function parseTagParam(value: string | null): number | undefined {
  if (value === null || !/^[1-9]\d*$/.test(value)) return undefined
  // 桁の多い数字は丸められて別のIDになるため、安全な整数に収まらなければ絞り込みなしにする
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : undefined
}

// 上限は API の q（Domain の検索キーワード）と同じ。String.length は UTF-16 単位で絵文字を 2 と数えるため使わない
const MAX_QUERY_LENGTH = 100

/** 検索キーワードが上限（前後の空白を除いてコードポイントで数える）を超えるかを判定する。 */
export function isQueryTooLong(value: string | undefined): boolean {
  return value !== undefined && Array.from(value.trim()).length > MAX_QUERY_LENGTH
}

/** 検索キーワードの前後の空白を除き、空ならキーワードなし（undefined）にする。 */
export function normalizeQuery(value: string): string | undefined {
  const trimmed = value.trim()
  return trimmed === '' ? undefined : trimmed
}
