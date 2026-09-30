import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '@/testing/server'
import { ApiError, apiMutator } from './mutator'

describe('apiMutator', () => {
  it('成功時はレスポンスボディをそのまま返す', async () => {
    server.use(http.get('/api/ping', () => HttpResponse.json({ ok: true })))

    await expect(apiMutator('/api/ping')).resolves.toEqual({ ok: true })
  })

  it('204 のときは undefined を返す', async () => {
    server.use(
      http.delete('/api/ping', () => new HttpResponse(null, { status: 204 })),
    )

    await expect(
      apiMutator('/api/ping', { method: 'DELETE' }),
    ).resolves.toBeUndefined()
  })

  it('エラー時は ErrorResponse の message を持つ ApiError を投げる', async () => {
    server.use(
      http.get('/api/ping', () =>
        HttpResponse.json({ message: 'not found' }, { status: 404 }),
      ),
    )

    const err = await apiMutator('/api/ping').catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ status: 404, message: 'not found' })
  })

  it('エラーボディに message が無ければステータス入りの既定文言にする', async () => {
    server.use(
      http.get('/api/ping', () => new HttpResponse('oops', { status: 500 })),
    )

    await expect(apiMutator('/api/ping')).rejects.toMatchObject({
      status: 500,
      message: 'API error: 500',
    })
  })

  it('ネットワークエラーは status 0 の ApiError に正規化する', async () => {
    server.use(http.get('/api/ping', () => HttpResponse.error()))

    await expect(apiMutator('/api/ping')).rejects.toMatchObject({ status: 0 })
  })

  it('成功ステータスでも不正な JSON なら ApiError を投げる', async () => {
    server.use(
      http.get(
        '/api/ping',
        () => new HttpResponse('not json', { status: 200 }),
      ),
    )

    await expect(apiMutator('/api/ping')).rejects.toBeInstanceOf(ApiError)
  })

  it('入力検証の 400 はフィールドごとの誤りを fieldErrors に持つ', async () => {
    server.use(
      http.post('/api/ping', () =>
        HttpResponse.json(
          {
            message: 'validation failed',
            errors: [
              { field: 'summary', rule: 'required' },
              { field: 'comment', rule: 'max' },
            ],
          },
          { status: 400 },
        ),
      ),
    )

    await expect(
      apiMutator('/api/ping', { method: 'POST' }),
    ).rejects.toMatchObject({
      status: 400,
      fieldErrors: [
        { field: 'summary', rule: 'required' },
        { field: 'comment', rule: 'max' },
      ],
    })
  })

  it('errors が無い・形が崩れた要素は fieldErrors に入れない', async () => {
    server.use(
      http.post('/api/ping', () =>
        HttpResponse.json(
          { message: 'bad', errors: [{ field: 'summary' }, 'x', null] },
          { status: 400 },
        ),
      ),
      http.get('/api/ping', () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
    )

    await expect(
      apiMutator('/api/ping', { method: 'POST' }),
    ).rejects.toMatchObject({ fieldErrors: [] })
    await expect(apiMutator('/api/ping')).rejects.toMatchObject({
      fieldErrors: [],
    })
  })

  it('ネットワークエラーの fieldErrors は空', async () => {
    server.use(http.get('/api/ping', () => HttpResponse.error()))

    await expect(apiMutator('/api/ping')).rejects.toMatchObject({
      fieldErrors: [],
    })
  })
})
