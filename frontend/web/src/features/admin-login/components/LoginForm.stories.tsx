import type { Meta, StoryObj } from '@storybook/react-vite'
import { LoginForm } from './LoginForm'

// 管理画面（自分だけ）のログインフォーム。ID とパスワードを入れて入る
const meta = {
  title: 'features/admin-login/LoginForm',
  component: LoginForm,
  args: {
    id: '',
    password: '',
    onIdChange: () => {},
    onPasswordChange: () => {},
    onSubmit: () => {},
    submitting: false,
  },
} satisfies Meta<typeof LoginForm>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Error: Story = {
  args: { id: 'admin', password: 'wrong', error: 'IDかパスワードが違います' },
}

export const Submitting: Story = {
  args: { id: 'admin', password: 'pw', submitting: true },
}
