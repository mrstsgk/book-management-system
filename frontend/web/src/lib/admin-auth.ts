// 管理画面の書き込みに付ける管理者トークン。ログイン画面を作らず環境変数から読む（ADR: 2026-09-30-admin-token-from-env）。
// mutator で全要求に付けないのは、公開画面の要求にトークンを載せないため
export function adminToken(): string {
  return (import.meta.env.VITE_ADMIN_TOKEN ?? '').trim()
}

export function adminRequest(): RequestInit {
  const token = adminToken()
  return token ? { headers: { Authorization: `Bearer ${token}` } } : {}
}
