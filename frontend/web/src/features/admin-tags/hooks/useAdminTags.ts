import {
  useDeleteTag,
  useSearchTags,
  useCountBooks,
  useRegisterTag,
  useRenameTag,
} from '@/api/generated/api'
import { ApiError } from '@/api/mutator'
import { useLoginRedirect } from '@/hooks/useLoginRedirect'
import { mergeTagCounts } from '@/features/admin-tags/utils/mergeTagCounts'
import { toFieldMessages } from '@/utils/serverFieldErrors'

type Outcome = { ok: true } | { ok: false; message: string; stale?: boolean }

const genericError = (action: string) =>
  `${action}できませんでした。時間をおいてもう一度お試しください。`

// 同じ名前のタグは 400 の errors[{field: 'name', rule: 'tagNameDuplicated'}] で返る。
// 409 は楽観ロック（version 不一致）だけなので、文言で区別しなくてよい
function nameError(e: unknown): string | undefined {
  if (!(e instanceof ApiError) || e.status !== 400) return undefined
  return toFieldMessages(e.fieldErrors).name
}

export function useAdminTags() {
  const redirectIfUnauthorized = useLoginRedirect()
  const tags = useSearchTags()
  const counts = useCountBooks()
  const addMutation = useRegisterTag()
  const renameMutation = useRenameTag()
  const deleteMutation = useDeleteTag()

  const refetchList = () => {
    void tags.refetch()
    void counts.refetch()
  }

  const addTag = async (name: string): Promise<Outcome> => {
    try {
      await addMutation.mutateAsync({ data: { name } })
      refetchList()
      return { ok: true }
    } catch (e) {
      if (redirectIfUnauthorized(e))
        return { ok: false, message: 'ログインしてください' }
      const message = nameError(e)
      if (message) return { ok: false, message }
      return { ok: false, message: genericError('追加') }
    }
  }

  const renameTag = async (
    id: number,
    version: number,
    name: string,
  ): Promise<Outcome> => {
    try {
      await renameMutation.mutateAsync({ id, data: { name, version } })
      refetchList()
      return { ok: true }
    } catch (e) {
      if (redirectIfUnauthorized(e))
        return { ok: false, message: 'ログインしてください' }
      const message = nameError(e)
      if (message) return { ok: false, message }
      if (e instanceof ApiError && e.status === 409) {
        refetchList()
        return {
          ok: false,
          stale: true,
          message:
            'このタグは、ほかの画面で先に更新されていました。最新の内容を読み込みました。',
        }
      }
      return { ok: false, message: genericError('保存') }
    }
  }

  const deleteTag = async (id: number): Promise<Outcome> => {
    try {
      await deleteMutation.mutateAsync({ id })
      refetchList()
      return { ok: true }
    } catch (e) {
      if (redirectIfUnauthorized(e))
        return { ok: false, message: 'ログインしてください' }
      return { ok: false, message: genericError('削除') }
    }
  }

  return {
    items: mergeTagCounts(tags.data?.items ?? [], counts.data?.items ?? []),
    isPending: tags.isPending || counts.isPending,
    isError: tags.isError || counts.isError,
    refetch: refetchList,
    addTag,
    renameTag,
    deleteTag,
  }
}
