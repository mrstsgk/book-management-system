export function barRatio(count: number, max: number): number {
  if (max <= 0) return 0
  return Math.round((count / max) * 100)
}
