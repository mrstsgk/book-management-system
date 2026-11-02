import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { getCheckSessionQueryKey } from '@/api/generated/api'
import { ApiError } from '@/api/mutator'

/** 表示中の読み取り（クエリ）がセッション切れ（401）になったら、今いた場所を添えてログイン画面へ送る。 */
export function useRedirectOnUnauthorizedRead() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { pathname } = useLocation()

  // 管理の読み取りはすべてセッションが要るため、画面ごとの hook ではなくクエリのキャッシュで一括して受ける
  // （書き込みは useLoginRedirect が個別に受ける。ミューテーションはキャッシュのイベントに載らない）
  useEffect(
    () =>
      queryClient.getQueryCache().subscribe((event) => {
        if (event.type !== 'updated' || event.action.type !== 'error') return
        const { error } = event.action
        if (error instanceof ApiError && error.status === 401) {
          // ログイン済みという古い確認結果を捨てる。残すとログイン画面が管理画面へ送り返してしまう
          queryClient.removeQueries({ queryKey: getCheckSessionQueryKey() })
          void navigate('/login', { state: { from: pathname } })
        }
      }),
    [queryClient, navigate, pathname],
  )
}
