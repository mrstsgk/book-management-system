import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { usePostApiAuthLogin } from '@/api/generated/api'
import { ApiError } from '@/api/mutator'

// 401 は ID とパスワードのどちらが違うかを出さない（バックエンドも区別して返さない）
function loginErrorMessage(error: unknown): string | undefined {
  if (!error) return undefined
  const status = error instanceof ApiError ? error.status : undefined
  if (status === 401) return 'IDかパスワードが違います'
  if (status === 429) return 'しばらく待ってからやり直してください'
  return 'ログインできませんでした。時間をおいてもう一度お試しください。'
}

export function useLogin() {
  const navigate = useNavigate()
  const location = useLocation()
  // useLoginRedirect / AdminGuard が state.from に元の場所を入れる
  const from = (location.state as { from?: string } | null)?.from
  const [id, setId] = useState('')
  const [password, setPassword] = useState('')
  const [localError, setLocalError] = useState<string | undefined>(undefined)
  const mutation = usePostApiAuthLogin()

  const submit = () => {
    if (mutation.isPending) return
    if (!id.trim() || !password) {
      setLocalError('IDとパスワードを入力してください')
      return
    }
    setLocalError(undefined)
    mutation.mutate(
      { data: { id, password } },
      {
        onSuccess: () => {
          navigate(from ?? '/admin', { replace: true })
        },
      },
    )
  }

  return {
    id,
    password,
    setId,
    setPassword,
    submit,
    submitting: mutation.isPending,
    error: localError ?? loginErrorMessage(mutation.error),
  }
}
