import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useGetApiCatalogIsbn,
  useGetApiTags,
  usePostApiBooks,
} from '@/api/generated/api'
import { ApiError } from '@/api/mutator'
import { useLoginRedirect } from '@/hooks/useLoginRedirect'
import {
  emptyBookFormValues,
  type BookFormErrors,
  type BookFormValues,
} from '@/features/admin-book-editor/types'
import { toFieldMessages } from '@/utils/serverFieldErrors'
import { validateBookForm } from '@/features/admin-book-editor/utils/validateBookForm'
import { visibleLocalErrors } from '@/features/admin-book-editor/utils/visibleErrors'

// 登録が失敗したとき、項目に対応しない誤りを画面上部の一般的な文言にする。
// 登録済み・カタログに無い・タグが無いは 400 の errors[{field, rule}] で項目ごとに返るので、
// ここに来る 400 は項目に対応しないもの、409 は一意制約に当たった想定外の競合
function generalRegisterMessage(error: ApiError | null): string | undefined {
  if (!error) return undefined
  if (error.fieldErrors.length > 0) return undefined
  return '登録できませんでした。時間をおいてもう一度お試しください。'
}

// ISBN を確かめた結果を、画面が出し分ける状態（見つからない・形式が不正・一時的な障害）に分類する。
// 400（形式・チェックディジットが不正）は入力の誤りであり再試行しても解消しないため、
// 一時的な障害（500・通信断など）とは別の文言にする
function catalogErrorState(error: ApiError | undefined) {
  return {
    isNotFound: error?.status === 404,
    isInvalid: error?.status === 400,
    isError:
      error !== undefined && error.status !== 404 && error.status !== 400,
  }
}

export function useBookRegistration() {
  const navigate = useNavigate()
  const redirectIfUnauthorized = useLoginRedirect()
  const [isbn, setIsbn] = useState('')
  const [confirmedIsbn, setConfirmedIsbn] = useState('')
  const [values, setValues] = useState<BookFormValues>(emptyBookFormValues)
  const [localErrors, setLocalErrors] = useState<BookFormErrors>({})
  const [registered, setRegistered] = useState(false)
  const submitting = useRef(false)

  const catalogQuery = useGetApiCatalogIsbn(confirmedIsbn || 'x', {
    query: { enabled: confirmedIsbn !== '' },
  })
  // ISBN の確認もログイン済みでないと通らない。画面を開いたままセッションが切れた場合、
  // 一般エラーで止めずに登録と同じくログイン画面へ送る。
  // redirectIfUnauthorized は現在地が変わるたびに別の関数になるため依存に入れない
  // （入れると、遷移後の現在地で二度目の遷移が走り state.from が /admin/login になる）
  const catalogError = catalogQuery.error
  useEffect(() => {
    if (catalogError) redirectIfUnauthorized(catalogError)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [catalogError])
  const tagsQuery = useGetApiTags()
  const registerMutation = usePostApiBooks()

  const confirm = () => {
    setConfirmedIsbn(isbn.trim())
  }

  const submit = () => {
    if (submitting.current) return
    const errors = validateBookForm(values)
    setLocalErrors(errors)
    if (Object.keys(errors).length > 0) return

    submitting.current = true
    registerMutation.mutate(
      {
        data: {
          // isbn（入力欄）ではなく confirmedIsbn を送る。確認後に入力欄だけ書き換えても
          // catalog.confirmed は true のままなので、確認していない ISBN で登録してしまわないため
          isbn: confirmedIsbn,
          summary: values.summary.trim(),
          comment: values.comment,
          rating: values.rating,
          tagIds: values.tagIds,
          titleOverride: values.titleOverride.trim() || undefined,
        },
      },
      {
        onSettled: () => {
          submitting.current = false
        },
        onError: redirectIfUnauthorized,
        onSuccess: () => {
          setRegistered(true)
          navigate('/admin', { state: { notice: '本を登録しました' } })
        },
      },
    )
  }

  // isbn はフォーム（BookFormValues）の外の入力欄なので、項目のエラーから分けて ISBN 欄の下に出す
  const { isbn: submittedIsbnError, ...serverErrors } = registerMutation.error
    ? toFieldMessages(registerMutation.error.fieldErrors)
    : {}
  // 送った ISBN に対する誤りなので、入力欄を別の ISBN に書き換えたら出さない。
  // mutation を reset しないのは、送った ISBN に戻したときに同じ誤りを出し直すため
  const isbnError =
    isbn.trim() === registerMutation.variables?.data.isbn
      ? submittedIsbnError
      : undefined
  const errors: BookFormErrors = {
    ...serverErrors,
    ...visibleLocalErrors(localErrors, values),
  }

  return {
    isbn,
    setIsbn,
    confirm,
    catalog: {
      data: catalogQuery.data,
      isLoading: catalogQuery.isFetching,
      ...catalogErrorState(catalogQuery.error ?? undefined),
      confirmed: confirmedIsbn !== '' && catalogQuery.isSuccess,
    },
    tags: tagsQuery.data?.items ?? [],
    values,
    setValues,
    errors,
    isbnError,
    generalError: generalRegisterMessage(registerMutation.error),
    submit,
    submitting: registerMutation.isPending,
    registered,
  }
}

export type BookRegistration = ReturnType<typeof useBookRegistration>
