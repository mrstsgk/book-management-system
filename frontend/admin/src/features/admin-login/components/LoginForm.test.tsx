import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { LoginForm } from './LoginForm'

function setupForm(over: Partial<Parameters<typeof LoginForm>[0]> = {}) {
  const props = {
    id: '',
    password: '',
    onIdChange: vi.fn(),
    onPasswordChange: vi.fn(),
    onSubmit: vi.fn(),
    submitting: false,
    ...over,
  }
  renderWithProviders(<LoginForm {...props} />)
  return props
}

describe('LoginForm', () => {
  it('ID とパスワードの欄と、ログインボタンを出す', () => {
    setupForm()
    const id = screen.getByLabelText('ID （必須）')
    expect(id).toBeVisible()
    expect(id).toHaveAttribute('autocomplete', 'username')
    const pw = screen.getByLabelText('パスワード （必須）')
    expect(pw).toHaveAttribute('type', 'password')
    expect(pw).toHaveAttribute('autocomplete', 'current-password')
    expect(screen.getByRole('button', { name: 'ログイン' })).toBeEnabled()
  })

  it('送信すると onSubmit だけが呼ばれる', async () => {
    const handlers = setupForm({ id: 'admin', password: 'pw' })
    await userEvent.click(screen.getByRole('button', { name: 'ログイン' }))
    expect(handlers.onSubmit).toHaveBeenCalledTimes(1)
  })

  it('入力すると onIdChange / onPasswordChange が呼ばれる', async () => {
    const handlers = setupForm()
    await userEvent.type(screen.getByLabelText('ID （必須）'), 'a')
    await userEvent.type(screen.getByLabelText('パスワード （必須）'), 'b')
    expect(handlers.onIdChange).toHaveBeenCalledWith('a')
    expect(handlers.onPasswordChange).toHaveBeenCalledWith('b')
  })

  it('エラーは role="alert" で出す', () => {
    setupForm({ error: 'IDかパスワードが違います' })
    expect(screen.getByRole('alert')).toHaveTextContent(
      'IDかパスワードが違います',
    )
  })

  it('送信中はボタンを押せない', () => {
    setupForm({ submitting: true })
    expect(screen.getByRole('button', { name: 'ログイン' })).toBeDisabled()
  })
})
