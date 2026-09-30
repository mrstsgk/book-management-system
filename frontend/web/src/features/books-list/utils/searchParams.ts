// URL の tag は正の整数だけを受け付ける。それ以外は API の 400 を画面に出さないよう絞り込みなしにする
export function parseTagParam(value: string | null): number | undefined {
  if (value === null || !/^[1-9]\d*$/.test(value)) return undefined
  // 桁の多い数字は丸められて別のIDになるため、安全な整数に収まらなければ絞り込みなしにする
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : undefined
}

export function normalizeQuery(value: string): string | undefined {
  const trimmed = value.trim()
  return trimmed === '' ? undefined : trimmed
}
