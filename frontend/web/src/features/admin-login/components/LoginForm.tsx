import { FormField } from '@/components/form/FormField'

type LoginFormProps = {
  id: string
  password: string
  onIdChange: (v: string) => void
  onPasswordChange: (v: string) => void
  onSubmit: () => void
  submitting: boolean
  error?: string
}

const inputClass = 'h-12 rounded-lg border border-ink-400 px-4 text-base'

export function LoginForm({
  id,
  password,
  onIdChange,
  onPasswordChange,
  onSubmit,
  submitting,
  error,
}: LoginFormProps) {
  return (
    <form
      className="card flex w-full max-w-sm flex-col gap-5 px-6 py-6"
      onSubmit={(e) => {
        e.preventDefault()
        onSubmit()
      }}
    >
      {error && (
        <p role="alert" className="text-sm font-bold text-brand-800">
          {error}
        </p>
      )}
      <FormField id="login-id" label="ID" required>
        <input
          id="login-id"
          className={inputClass}
          value={id}
          autoComplete="username"
          onChange={(e) => onIdChange(e.target.value)}
        />
      </FormField>
      <FormField id="login-password" label="パスワード" required>
        <input
          id="login-password"
          type="password"
          className={inputClass}
          value={password}
          autoComplete="current-password"
          onChange={(e) => onPasswordChange(e.target.value)}
        />
      </FormField>
      <button
        type="submit"
        disabled={submitting}
        className="h-12 rounded-lg bg-brand-700 px-6 font-bold text-white disabled:cursor-not-allowed disabled:bg-ink-400"
      >
        ログイン
      </button>
    </form>
  )
}
