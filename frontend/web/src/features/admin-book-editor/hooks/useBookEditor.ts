import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useDeleteApiBooksId,
  useGetApiBooksId,
  useGetApiTags,
  usePutApiBooksId,
} from '@/api/generated/api'
import type { BookResponse, TagResponse } from '@/api/generated/api.schemas'
import type { BookFormErrors, BookFormValues } from '../types'
import { parseBookId } from '../utils/parseBookId'
import { toFieldMessages } from '../utils/serverFieldErrors'
import { tagIdsByName } from '../utils/tagIds'
import { validateBookForm } from '../utils/validateBookForm'
import { visibleLocalErrors } from '../utils/visibleErrors'

// 取得した本とタグ一覧から、フォームの初期値を作る（タグは名前からIDに引き直す）
function buildFormValues(
  book: BookResponse,
  tags: TagResponse[],
): BookFormValues {
  return {
    titleOverride: book.titleOverride ?? '',
    summary: book.summary ?? '',
    comment: book.comment ?? '',
    rating: book.rating ?? 0,
    tagIds: tagIdsByName(book.tags ?? [], tags),
  }
}

// ID が無い、または 404 なら見つからない扱い
function isNotFound(
  id: number | undefined,
  bookStatus: number | undefined,
): boolean {
  return id === undefined || bookStatus === 404
}

// 見つからない扱いでないときだけ、本とタグ一覧のどちらかの状態を見る
function isBusyState(notFound: boolean, a: boolean, b: boolean): boolean {
  return !notFound && (a || b)
}

// 保存の 400 のフィールドエラーを画面の文言に変える。エラーが無ければ空
function serverFieldErrorsOf(
  error: { fieldErrors: Parameters<typeof toFieldMessages>[0] } | null,
): BookFormErrors {
  return error ? toFieldMessages(error.fieldErrors) : {}
}

export function useBookEditor(rawId: string | undefined) {
  const navigate = useNavigate()
  const id = parseBookId(rawId)

  const bookQuery = useGetApiBooksId(id ?? 0, {
    query: { enabled: id !== undefined },
  })
  const tagsQuery = useGetApiTags()
  const updateMutation = usePutApiBooksId()
  const deleteMutation = useDeleteApiBooksId()

  const [values, setValues] = useState<BookFormValues | null>(null)
  // 保存に使う version。編集開始時（またはreloadLatest時）に値とあわせて固定する。
  // bookQuery.data.version を送信時に直接読むと、reloadLatestを経ない裏の取り直し
  // （再接続時の自動再取得など）でversionだけ進み、古いフォーム値のまま新しいversionを
  // 送って409にならず、他の変更を無自覚に上書きしてしまう
  const [editingVersion, setEditingVersion] = useState<number | null>(null)
  const [localErrors, setLocalErrors] = useState<BookFormErrors>({})
  const [conflict, setConflict] = useState(false)
  const [saved, setSaved] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteError, setDeleteError] = useState<string | undefined>(undefined)
  const submitting = useRef(false)
  const deleting = useRef(false)

  const tags = tagsQuery.data?.items ?? []

  // 本とタグ一覧が揃ったら初期値を作る。一度作ったら、ユーザーの入力を上書きしない
  // （reloadLatest が明示的に null に戻したときだけ、ここでもう一度作り直す）
  useEffect(() => {
    if (values !== null) return
    if (!bookQuery.data || !tagsQuery.data) return
    setValues(buildFormValues(bookQuery.data, tags))
    setEditingVersion(bookQuery.data.version ?? null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bookQuery.data, tagsQuery.data, values])

  const notFound = isNotFound(id, bookQuery.error?.status)

  const submit = () => {
    if (submitting.current || !values || id === undefined) return
    const errors = validateBookForm(values)
    setLocalErrors(errors)
    if (Object.keys(errors).length > 0) return
    if (editingVersion === null) return

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
          version: editingVersion,
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

  // 取り直した本を直接使って値とversionを作る。setValues(null) → useEffect の順に
  // 任せると、refetch が終わる前の古い bookQuery.data で一瞬埋め直されてしまう競合があるため
  const reloadLatest = async () => {
    setConflict(false)
    const { data: latest } = await bookQuery.refetch()
    if (!latest) return
    setValues(buildFormValues(latest, tags))
    setEditingVersion(latest.version ?? null)
  }

  const openDelete = () => {
    setDeleteError(undefined)
    setDeleteOpen(true)
  }
  const closeDelete = () => setDeleteOpen(false)

  const confirmDelete = () => {
    if (deleting.current || id === undefined) return
    deleting.current = true
    setDeleteError(undefined)
    const title = bookQuery.data?.title ?? ''
    deleteMutation.mutate(
      { id },
      {
        onSettled: () => {
          deleting.current = false
        },
        onSuccess: () => {
          navigate('/admin', {
            state: { notice: `『${title}』を削除しました` },
          })
        },
        onError: () => {
          setDeleteError(
            '削除できませんでした。時間をおいてもう一度お試しください。',
          )
        },
      },
    )
  }

  const serverErrors = serverFieldErrorsOf(updateMutation.error)
  const errors: BookFormErrors = {
    ...serverErrors,
    ...(values ? visibleLocalErrors(localErrors, values) : localErrors),
  }

  const retry = () => {
    void bookQuery.refetch()
    void tagsQuery.refetch()
  }

  return {
    notFound,
    isLoading: isBusyState(notFound, bookQuery.isPending, tagsQuery.isPending),
    isError: isBusyState(notFound, bookQuery.isError, tagsQuery.isError),
    retry,
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
    deleteError,
    openDelete,
    closeDelete,
    confirmDelete,
    deleting: deleteMutation.isPending,
  }
}
