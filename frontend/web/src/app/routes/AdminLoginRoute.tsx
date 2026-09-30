import { LoginForm } from '@/features/admin-login/components/LoginForm'
import { useLogin } from '@/features/admin-login/hooks/useLogin'

export function AdminLoginRoute() {
  const login = useLogin()
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
