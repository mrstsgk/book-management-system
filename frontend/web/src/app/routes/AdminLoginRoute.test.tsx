import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { AppRoutes } from '@/app/router'
import { renderWithProviders } from '@/testing/render'
import { server } from '@/testing/server'

describe('AdminLoginRoute', () => {
  it('/admin/login で ID とパスワードを入れてログインすると /admin の管理画面へ移る', async () => {
    server.use(
      http.post(
        '*/api/auth/login',
        () => new HttpResponse(null, { status: 204 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    expect(
      screen.getByRole('heading', { name: '管理画面にログイン' }),
    ).toBeVisible()
    await userEvent.type(screen.getByLabelText('ID （必須）'), 'admin')
    await userEvent.type(screen.getByLabelText('パスワード （必須）'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'ログイン' }))

    expect(
      await screen.findByRole('navigation', { name: '管理メニュー' }),
    ).toBeVisible()
  })

  it('間違えると文言を出し、入力は残る', async () => {
    server.use(
      http.post('*/api/auth/login', () =>
        HttpResponse.json({ message: 'unauthorized' }, { status: 401 }),
      ),
    )
    renderWithProviders(<AppRoutes />, { route: '/admin/login' })

    await userEvent.type(screen.getByLabelText('ID （必須）'), 'admin')
    await userEvent.type(screen.getByLabelText('パスワード （必須）'), 'bad')
    await userEvent.click(screen.getByRole('button', { name: 'ログイン' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'IDかパスワードが違います',
    )
    expect(screen.getByLabelText('ID （必須）')).toHaveValue('admin')
  })
})
