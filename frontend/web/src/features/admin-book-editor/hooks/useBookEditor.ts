import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useDeleteApiBooksId,
  useGetApiBooksId,
  useGetApiTags,
  usePutApiBooksId,
} from '@/api/generated/api'
import { adminRequest } from '@/lib/admin-auth'
import type { BookFormErrors, BookFormValues } from '../types'
import { toFieldMessages } from '../utils/serverFieldErrors'
import { tagIdsByName } from '../utils/tagIds'
import { validateBookForm } from '../utils/validateBookForm'

// 本の ID。桁あふれで別の本を指してしまわないよう安全な整数に限る（parseBookId と同じ考え方）
function parseId(raw: string | undefined): number | undefined {
  if (raw === undefined || !/^\d+$/.test(raw)) return undefined
  const id = Number(raw)
  return Number.isSafeInteger(id) && id >= 1 ? id : undefined
}

export function useBookEditor(rawId: string | undefined) {
  const navigate = useNavigate()
  const id = parseId(rawId)

  const bookQuery = useGetApiBooksId(id ?? 0, {
    query: { enabled: id !== undefined },
    request: adminRequest(),
  })
  const tagsQuery = useGetApiTags({ request: adminRequest() })
  const updateMutation = usePutApiBooksId({ request: adminRequest() })
  const deleteMutation = useDeleteApiBooksId({ request: adminRequest() })

  const [values, setValues] = useState<BookFormValues | null>(null)
  const [localErrors, setLocalErrors] = useState<BookFormErrors>({})
  const [conflict, setConflict] = useState(false)
  const [saved, setSaved] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const submitting = useRef(false)
  const deleting = useRef(false)

  const tags = tagsQuery.data?.items ?? []

  // 本とタグ一覧が揃ったら初期値を作る。一度作ったら、ユーザーの入力を上書きしない
  // （reloadLatest が明示的に null に戻したときだけ、ここでもう一度作り直す）
  useEffect(() => {
    if (values !== null) return
    if (!bookQuery.data || !tagsQuery.data) return
    setValues({
      titleOverride: bookQuery.data.titleOverride ?? '',
      summary: bookQuery.data.summary ?? '',
      comment: bookQuery.data.comment ?? '',
      rating: bookQuery.data.rating ?? 0,
      tagIds: tagIdsByName(bookQuery.data.tags ?? [], tags),
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bookQuery.data, tagsQuery.data, values])

  const notFound = id === undefined || bookQuery.error?.status === 404

  const submit = () => {
    if (submitting.current || !values || id === undefined) return
    const errors = validateBookForm(values)
    setLocalErrors(errors)
    if (Object.keys(errors).length > 0) return
    const version = bookQuery.data?.version
    if (version === undefined) return

    submitting.current = true
    updateMutation.mutate(
      {
        id,
        data: {
          summary: values.summary.trim(),
          comment: values.comment,
          rating: values.rating,
          tagIds: values.tagIds,
          titleOverride: values.titleOverride.trim() || undefined,
          version,
        },
      },
      {
        onSettled: () => {
          submitting.current = false
        },
        onSuccess: () => {
          setSaved(true)
          navigate('/admin', { state: { notice: '本を更新しました' } })
        },
        onError: (err) => {
          if (err.status === 409) setConflict(true)
        },
      },
    )
  }

  // 取り直した本を直接使って値を作る。setValues(null) → useEffect の順に任せると、
  // refetch が終わる前の古い bookQuery.data で一瞬埋め直されてしまう競合があるため
  const reloadLatest = async () => {
    setConflict(false)
    const { data: latest } = await bookQuery.refetch()
    if (!latest) return
    setValues({
      titleOverride: latest.titleOverride ?? '',
      summary: latest.summary ?? '',
      comment: latest.comment ?? '',
      rating: latest.rating ?? 0,
      tagIds: tagIdsByName(latest.tags ?? [], tags),
    })
  }

  const openDelete = () => setDeleteOpen(true)
  const closeDelete = () => setDeleteOpen(false)

  const confirmDelete = () => {
    if (deleting.current || id === undefined) return
    deleting.current = true
    const title = bookQuery.data?.title ?? ''
    deleteMutation.mutate(
      { id },
      {
        onSettled: () => {
          deleting.current = false
        },
        onSuccess: () => {
          navigate('/admin', { state: { notice: `『${title}』を削除しました` } })
        },
      },
    )
  }

  const serverErrors = updateMutation.error
    ? toFieldMessages(updateMutation.error.fieldErrors)
    : {}
  const errors: BookFormErrors = { ...serverErrors, ...localErrors }

  return {
    notFound,
    isLoading: !notFound && (bookQuery.isPending || tagsQuery.isPending),
    isError: !notFound && (bookQuery.isError || tagsQuery.isError),
    retry: () => {
      void bookQuery.refetch()
      void tagsQuery.refetch()
    },
    book: bookQuery.data,
    values,
    setValues,
    tags,
    errors,
    conflict,
    reloadLatest,
    submit,
    submitting: updateMutation.isPending,
    saved,
    deleteOpen,
    openDelete,
    closeDelete,
    confirmDelete,
    deleting: deleteMutation.isPending,
  }
}
