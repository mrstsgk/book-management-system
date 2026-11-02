// Number() は '1.5' や '1e3'・前後の空白も数値にしてしまうため、数字だけの文字列に限る。
// 桁が多すぎて精度を失う値は別の本の ID になりうるので、安全な整数に限る
export function parseBookId(raw: string | undefined): number | undefined {
  if (raw === undefined || !/^\d+$/.test(raw)) return undefined
  const id = Number(raw)
  return Number.isSafeInteger(id) && id >= 1 ? id : undefined
}
