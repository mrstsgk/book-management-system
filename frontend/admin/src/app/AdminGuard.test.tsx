import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

function serveSession(status: number) {
  server.use(
    http.get('*/api/admin/auth/session', () =>
      status === 204
        ? new HttpResponse(null, { status: 204 })
        : HttpResponse.json({ message: 'x' }, { status }),
    ),
  )
}

describe('AdminGuard', () => {
  it('確認中は読み込み中を出す', () => {
    serveSession(204)
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
  })

  it('ログイン済みなら管理画面を出す', async () => {
    serveSession(204)
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    expect(
      await screen.findByRole('heading', { name: '分野タグ' }),
    ).toBeVisible()
  })

  it.each([401, 500])('%s なら /login へ送る', async (status) => {
    serveSession(status)
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('navigation', { name: '管理メニュー' }),
    ).not.toBeInTheDocument()
  })

  it('表示中の読み取りがセッション切れ（401）になったら、元の場所を添えて /login へ送る', async () => {
    // 最初の確認ではログイン済み、その後にセッションが切れた状況
    let sessionChecks = 0
    server.use(
      http.get('*/api/admin/auth/session', () => {
        sessionChecks += 1
        return sessionChecks === 1
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ message: 'unauthenticated' }, { status: 401 })
      }),
      http.get('*/api/admin/tags', () =>
        HttpResponse.json({ message: 'unauthenticated' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
  })

  it('読み取りが 401 以外で失敗しても /login へは送らない', async () => {
    serveSession(204)
    server.use(
      http.get('*/api/admin/tags', () =>
        HttpResponse.json({ message: 'internal' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    expect(
      await screen.findByRole('heading', { name: '分野タグ' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: '管理画面にログイン' }),
    ).not.toBeInTheDocument()
  })

  it('ログインしたら元の管理画面へ戻る', async () => {
    serveSession(401)
    const user = userEvent.setup()
    renderWithProviders(<AppRoutes />, { route: '/tags' })
    await screen.findByRole('heading', { name: '管理画面にログイン' })

    server.use(
      http.post(
        '*/api/admin/auth/login',
        () => new HttpResponse(null, { status: 204 }),
      ),
    )
    serveSession(204)
    await user.type(screen.getByLabelText('ID （必須）'), 'admin')
    await user.type(screen.getByLabelText('パスワード （必須）'), 'pw')
    await user.click(screen.getByRole('button', { name: 'ログイン' }))

    expect(
      await screen.findByRole('heading', { name: '分野タグ' }),
    ).toBeVisible()
  })
})
