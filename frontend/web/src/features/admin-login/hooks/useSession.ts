import { useQuery } from '@tanstack/react-query'
import {
  getApiAuthSession,
  getGetApiAuthSessionQueryKey,
} from '@/api/generated/api'

export type SessionStatus = 'loading' | 'ok' | 'unauthorized'

// 管理画面を開くたびに 1 回だけ確かめる（再取得は要らない。書き込みで 401 になれば useLoginRedirect が送る）。
// 204 以外はすべて「ログインし直す」扱い: サーバーが落ちていればログインも失敗し、その文言で伝わるので、
// ここにエラー表示と再試行は持たない。
// 生成フックを使わない理由: 204 はボディ無しで undefined が返り、TanStack Query が「データ undefined」でエラー扱いにするため、true に置き換える
export function useSession() {
  const query = useQuery({
    queryKey: getGetApiAuthSessionQueryKey(),
    queryFn: async ({ signal }) => {
      await getApiAuthSession({ signal })
      return true
    },
    retry: false,
    staleTime: Infinity,
  })
  const status: SessionStatus = query.isPending
    ? 'loading'
    : query.isSuccess
      ? 'ok'
      : 'unauthorized'
  return { status }
}
