// Orval が生成する全リクエストはこの関数を通る（orval.config.ts の override.mutator）

/** Empty in local dev (Vite proxies /api). Set VITE_API_BASE_URL for deployed API. */
const baseUrl = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/+$/, '')

export interface FieldErrorDetail {
  field: string
  rule: string
  params?: Record<string, unknown>
}

export class ApiError extends Error {
  readonly status: number
  // 入力検証・相関ルールの 400 のときだけ中身がある（field は JSON 名）
  readonly fieldErrors: FieldErrorDetail[]
  // 業務ルールの違反のときだけ付くコード（x-error-codes）と、文言に差し込む値
  readonly code: string | undefined
  readonly params: Record<string, unknown>

  constructor(
    status: number,
    message: string,
    fieldErrors: FieldErrorDetail[] = [],
    code?: string,
    params: Record<string, unknown> = {},
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fieldErrors = fieldErrors
    this.code = code
    this.params = params
  }
}

// Orval はこの型を生成フックの TError に使う（mutator が投げるのは常に ApiError）
export type ErrorType<_Body> = ApiError

// エラーのボディは ErrorResponse（openapi/common/schemas/ErrorResponse.yaml）。形が崩れていても落とさない
function extractMessage(payload: unknown): string | undefined {
  if (typeof payload !== 'object' || payload === null) return undefined
  const { message } = payload as { message?: unknown }
  return typeof message === 'string' ? message : undefined
}

function extractCode(payload: unknown): string | undefined {
  if (typeof payload !== 'object' || payload === null) return undefined
  const { code } = payload as { code?: unknown }
  return typeof code === 'string' ? code : undefined
}

function extractParams(payload: unknown): Record<string, unknown> {
  if (typeof payload !== 'object' || payload === null) return {}
  const { params } = payload as { params?: unknown }
  return typeof params === 'object' && params !== null
    ? (params as Record<string, unknown>)
    : {}
}

// 入力検証・相関ルールの 400 は errors[{field, rule}] を持つ。形が崩れた要素は捨てる
function extractFieldErrors(payload: unknown): FieldErrorDetail[] {
  if (typeof payload !== 'object' || payload === null) return []
  const { errors } = payload as { errors?: unknown }
  if (!Array.isArray(errors)) return []
  return errors.filter(
    (e): e is FieldErrorDetail =>
      typeof e === 'object' &&
      e !== null &&
      typeof (e as { field?: unknown }).field === 'string' &&
      typeof (e as { rule?: unknown }).rule === 'string',
  )
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
      extractFieldErrors(payload),
      extractCode(payload),
      extractParams(payload),
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
