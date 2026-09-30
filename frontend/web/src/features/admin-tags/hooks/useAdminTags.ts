import {
  useDeleteApiTagsId,
  useGetApiTags,
  useGetApiTagsCounts,
  usePostApiTags,
  usePutApiTagsId,
} from '@/api/generated/api'
import { ApiError } from '@/api/mutator'
import { adminRequest } from '@/lib/admin-auth'
import { mergeTagCounts } from '../utils/mergeTagCounts'

type Outcome = { ok: true } | { ok: false; message: string; stale?: boolean }

const genericError = (action: string) =>
  `${action}できませんでした。時間をおいてもう一度お試しください。`

export function useAdminTags() {
  const tags = useGetApiTags()
  const counts = useGetApiTagsCounts()
  const addMutation = usePostApiTags({ request: adminRequest() })
  const renameMutation = usePutApiTagsId({ request: adminRequest() })
  const deleteMutation = useDeleteApiTagsId({ request: adminRequest() })

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
      if (e instanceof ApiError && e.status === 409) {
        return { ok: false, message: `「${name}」というタグはすでにあります` }
      }
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
    } catch {
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
