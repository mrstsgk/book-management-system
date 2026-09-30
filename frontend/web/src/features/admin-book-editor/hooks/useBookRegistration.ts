import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useGetApiCatalogIsbn,
  useGetApiTags,
  usePostApiBooks,
} from '@/api/generated/api'
import { ApiError } from '@/api/mutator'
import { adminRequest } from '@/lib/admin-auth'
import { emptyBookFormValues, type BookFormErrors, type BookFormValues } from '../types'
import { toFieldMessages } from '../utils/serverFieldErrors'
import { validateBookForm } from '../utils/validateBookForm'

// 登録が失敗したとき、フィールドの誤り以外（カタログに無い・二重登録など）の一般的な文言にする
function generalRegisterMessage(error: ApiError | null): string | undefined {
  if (!error) return undefined
  if (error.fieldErrors.length > 0) return undefined
  if (error.status === 409) return 'この ISBN の本はすでに登録されています。'
  if (error.status === 400)
    return 'この ISBN の本は外部カタログに見つかりませんでした。'
  return '登録できませんでした。時間をおいてもう一度お試しください。'
}

export function useBookRegistration() {
  const navigate = useNavigate()
  const [isbn, setIsbn] = useState('')
  const [confirmedIsbn, setConfirmedIsbn] = useState('')
  const [values, setValues] = useState<BookFormValues>(emptyBookFormValues)
  const [localErrors, setLocalErrors] = useState<BookFormErrors>({})
  const [registered, setRegistered] = useState(false)
  const submitting = useRef(false)

  const catalogQuery = useGetApiCatalogIsbn(confirmedIsbn || 'x', {
    query: { enabled: confirmedIsbn !== '' },
    request: adminRequest(),
  })
  const tagsQuery = useGetApiTags({ request: adminRequest() })
  const registerMutation = usePostApiBooks({ request: adminRequest() })

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
          isbn: isbn.trim(),
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
        onSuccess: () => {
          setRegistered(true)
          navigate('/admin', { state: { notice: '本を登録しました' } })
        },
      },
    )
  }

  const serverErrors = registerMutation.error
    ? toFieldMessages(registerMutation.error.fieldErrors)
    : {}
  const errors: BookFormErrors = { ...serverErrors, ...localErrors }

  return {
    isbn,
    setIsbn,
    confirm,
    catalog: {
      data: catalogQuery.data,
      isLoading: catalogQuery.isFetching,
      isNotFound: catalogQuery.error?.status === 404,
      isError: catalogQuery.isError && catalogQuery.error?.status !== 404,
      confirmed: confirmedIsbn !== '' && catalogQuery.isSuccess,
    },
    tags: tagsQuery.data?.items ?? [],
    values,
    setValues,
    errors,
    generalError: generalRegisterMessage(registerMutation.error),
    submit,
    submitting: registerMutation.isPending,
    registered,
  }
}
