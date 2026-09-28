import { getBookManagementSystemAPIMock } from '@/api/generated/api.msw'

// 既定は Orval 生成（faker のランダム値）。具体値を検証するテストは
// server.use(get...MockHandler(fixture)) で上書きする
export const handlers = getBookManagementSystemAPIMock()
