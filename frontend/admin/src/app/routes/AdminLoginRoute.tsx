import { Navigate, useLocation } from 'react-router-dom'
import { LoadingState } from '@/components/states/LoadingState'
import { LoginForm } from '@/features/admin-login/components/LoginForm'
import { useLoginForm } from '@/features/admin-login/hooks/useLoginForm'
import { useSession } from '@/features/admin-login/hooks/useSession'

export function AdminLoginRoute() {
  const { status } = useSession()
  const location = useLocation()
  const login = useLoginForm()

  // 確かめ終わるまでフォームを出さない（ログイン済みのときにフォームが一瞬見えてから移るのを避けるため）
  if (status === 'loading') return <LoadingState />
  if (status === 'ok') {
    // useLoginRedirect / AdminGuard が state.from に元の場所を入れる
    const from = (location.state as { from?: string } | null)?.from
    return <Navigate to={from ?? '/books'} replace />
  }
  return (
    <main className="flex min-h-full flex-col items-center justify-center gap-8 px-4 py-16">
      <h1 className="text-[28px] font-extrabold">管理画面にログイン</h1>
      <LoginForm
        id={login.id}
        password={login.password}
        onIdChange={login.setId}
        onPasswordChange={login.setPassword}
        onSubmit={login.submit}
        submitting={login.submitting}
        error={login.error}
      />
    </main>
  )
}
