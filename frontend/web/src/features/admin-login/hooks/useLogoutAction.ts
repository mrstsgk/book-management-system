import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  getCheckSessionQueryKey,
  useLogout,
} from '@/api/generated/api'

// 成功したときだけログイン画面へ送る: 失敗時はサーバー側のセッションも Cookie も残っている
// （バックエンドは消せなかったとき Cookie を消さずに 500 を返す）ため、画面だけ抜けたように
// 見せると、ログアウトしたつもりでセッションが生きたままになる。失敗はその場で伝えて再試行させる
/** ログアウトし、成功したらログイン画面へ送る操作と、その状態を提供する。 */
export function useLogoutAction() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const mutation = useLogout()
  const logout = () => {
    if (mutation.isPending) return
    mutation.mutate(undefined, {
      onSuccess: () => {
        // セッション確認の結果（ログイン済み）を捨てる。残すとログイン画面がそれを見て管理画面へ戻してしまう
        queryClient.removeQueries({ queryKey: getCheckSessionQueryKey() })
        navigate('/admin/login', { replace: true })
      },
    })
  }
  return { logout, pending: mutation.isPending, failed: mutation.isError }
}
