import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

// 未ログイン（セッション確認が 401）にする。既定のハンドラはログイン済み（204）
function loggedOut() {
  server.use(
    http.get('*/api/admin/auth/session', () =>
      HttpResponse.json({ message: 'unauthorized' }, { status: 401 }),
    ),
  )
}

describe('AdminLoginRoute', () => {
  it('ログイン済みで開くと、フォームを出さずに管理画面へ移る', async () => {
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(
      await screen.findByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
    expect(
      screen.queryByRole('heading', { name: '管理画面にログイン' }),
    ).not.toBeInTheDocument()
  })

  it('ログイン済みで元の場所（from）付きで開くと、その場所へ移る', async () => {
    renderWithProviders(<AppRoutes />, {
      route: { pathname: '/admin/login', state: { from: '/admin/tags' } },
    })

    expect(
      await screen.findByRole('heading', { level: 1, name: '分野タグ' }),
    ).toBeVisible()
  })

  it('セッションを確かめている間は読み込み中を出し、フォームはまだ出さない', () => {
    server.use(
      http.get('*/api/admin/auth/session', () => new Promise(() => {})),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(
      screen.queryByRole('heading', { name: '管理画面にログイン' }),
    ).not.toBeInTheDocument()
  })

  it('セッションの確認に失敗（サーバーの障害）したら、ログインのフォームを出す', async () => {
    server.use(
      http.get('*/api/admin/auth/session', () =>
        HttpResponse.json({ message: 'internal' }, { status: 500 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
  })

  it('/admin/login で ID とパスワードを入れてログインすると /admin の管理画面へ移る', async () => {
    loggedOut()
    server.use(
      http.post(
        '*/api/admin/auth/login',
        () => new HttpResponse(null, { status: 204 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(
      await screen.findByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    // ログインした後はセッション確認も通る
    server.use(
      http.get(
        '*/api/admin/auth/session',
        () => new HttpResponse(null, { status: 204 }),
      ),
    )
    await userEvent.type(screen.getByLabelText('ID （必須）'), 'admin')
    await userEvent.type(screen.getByLabelText('パスワード （必須）'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'ログイン' }))

    expect(
      await screen.findByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
  })

  it('間違えると文言を出し、入力は残る', async () => {
    loggedOut()
    server.use(
      http.post('*/api/admin/auth/login', () =>
        HttpResponse.json({ message: 'unauthorized' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    await userEvent.type(await screen.findByLabelText('ID （必須）'), 'admin')
    await userEvent.type(screen.getByLabelText('パスワード （必須）'), 'bad')
    await userEvent.click(screen.getByRole('button', { name: 'ログイン' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'IDかパスワードが違います',
    )
    expect(screen.getByLabelText('ID （必須）')).toHaveValue('admin')
  })
})
