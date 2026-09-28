import type { RequestHandler } from 'msw'

// Orval は paths が空の OpenAPI からは api.msw.ts を生成しない。エンドポイント追加後は
// getBookManagementSystemAPIMock()（生成 MSW ハンドラ）をここで既定として返す
export const handlers: RequestHandler[] = []
