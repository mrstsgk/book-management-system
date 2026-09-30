import { useNavigate } from 'react-router-dom'
import { usePostApiAuthLogout } from '@/api/generated/api'

// 失敗してもログイン画面へ送る: サーバー側が消せなくてもセッションは期限で切れ、画面上は「抜けた」状態にしたい
export function useLogout() {
  const navigate = useNavigate()
  const mutation = usePostApiAuthLogout()
  const logout = () => {
    if (mutation.isPending) return
    mutation.mutate(undefined, {
      onSettled: () => navigate('/admin/login', { replace: true }),
    })
  }
  return { logout, pending: mutation.isPending }
}
