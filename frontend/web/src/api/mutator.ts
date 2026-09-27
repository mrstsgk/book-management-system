// Orval が生成する全リクエストはこの関数を通る（orval.config.ts の override.mutator）

/** Empty in local dev (Vite proxies /api). Set VITE_API_BASE_URL for deployed API. */
const baseUrl = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/+$/, '')

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// Orval はこの型を生成フックの TError に使う（mutator が投げるのは常に ApiError）
export type ErrorType<_Body> = ApiError

// Echo は業務エラー（HandleError）も echo.HTTPError も { message } で返す
function extractMessage(payload: unknown): string | undefined {
  if (typeof payload !== 'object' || payload === null) return undefined
  const { message } = payload as { message?: unknown }
  return typeof message === 'string' ? message : undefined
}

export async function apiMutator<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${baseUrl}${url}`, options)
  } catch (e) {
    // fetch 自体の失敗（オフライン等）は TypeError で reject される。ApiError に正規化する
    throw new ApiError(0, e instanceof Error ? e.message : 'network error')
  }

  if (!response.ok) {
    const payload: unknown = await response.json().catch(() => null)
    throw new ApiError(
      response.status,
      extractMessage(payload) ?? `API error: ${response.status}`,
    )
  }

  // 204 No Content（削除など）はボディが無く JSON パースできない
  if (response.status === 204) return undefined as T
  try {
    return (await response.json()) as T
  } catch (e) {
    // 成功ステータスでも不正な JSON は SyntaxError で reject される。同様に正規化する
    throw new ApiError(
      response.status,
      e instanceof Error ? e.message : 'invalid response body',
    )
  }
}
