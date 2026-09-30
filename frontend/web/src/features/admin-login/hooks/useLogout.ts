import { useNavigate } from 'react-router-dom'
import { usePostApiAuthLogout } from '@/api/generated/api'

// 成功したときだけログイン画面へ送る: 失敗時はサーバー側のセッションも Cookie も残っている
// （バックエンドは消せなかったとき Cookie を消さずに 500 を返す）ため、画面だけ抜けたように
// 見せると、ログアウトしたつもりでセッションが生きたままになる。失敗はその場で伝えて再試行させる
export function useLogout() {
  const navigate = useNavigate()
  const mutation = usePostApiAuthLogout()
  const logout = () => {
    if (mutation.isPending) return
    mutation.mutate(undefined, {
      onSuccess: () => navigate('/admin/login', { replace: true }),
    })
  }
  return { logout, pending: mutation.isPending, failed: mutation.isError }
}
