import { useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '@/api/mutator'

// 書き込み中にセッションが切れた（401）ら、今いた場所を添えてログイン画面へ送る。
// mutator に共通の 401 処理を置かないのは、公開画面の要求が 401 を受けることは無く、
// 管理画面の 3 つの hook に 1 行ずつ書く方が「どこで遷移するか」が読めるため
export function useLoginRedirect() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  return (error: unknown): boolean => {
    if (!(error instanceof ApiError) || error.status !== 401) return false
    navigate('/admin/login', { state: { from: pathname } })
    return true
  }
}
