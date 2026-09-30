import { http, HttpResponse } from 'msw'
import { getBookManagementSystemAPIMock } from '@/api/generated/api.msw'

// 既定は Orval 生成（faker のランダム値）。具体値を検証するテストは
// server.use(get...MockHandler(fixture)) で上書きする。
// セッション確認だけは「ログイン済み（204）」を先頭で固定する（管理画面のテストの大半は
// ログイン済み前提で、未ログインの分岐は AdminGuard のテストが server.use で上書きする）
export const handlers = [
  http.get('*/api/auth/session', () => new HttpResponse(null, { status: 204 })),
  ...getBookManagementSystemAPIMock(),
]
