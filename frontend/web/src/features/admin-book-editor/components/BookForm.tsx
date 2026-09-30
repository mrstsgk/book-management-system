import { describedBy } from '@/components/form/describedBy'
import { ErrorSummary } from '@/components/form/ErrorSummary'
import { FormField } from '@/components/form/FormField'
import type { TagResponse } from '@/api/generated/api.schemas'
import type { BookFormErrors, BookFormValues } from '../types'

const RATING_LABELS: Record<number, string> = {
  1: '★',
  2: '★★',
  3: '★★★',
  4: '★★★★',
  5: '★★★★★',
}
const FIELD_LABELS: Record<keyof BookFormValues, string> = {
  titleOverride: '書名の上書き',
  summary: '一言まとめ',
  comment: '感想',
  rating: '評価',
  tagIds: '分野タグ',
}

type BookFormProps = {
  values: BookFormValues
  onChange: (values: BookFormValues) => void
  errors: BookFormErrors
  tags: TagResponse[]
  submitLabel: string
  onSubmit: () => void
  submitting: boolean
  disabled?: boolean
}

export function BookForm({
  values,
  onChange,
  errors,
  tags,
  submitLabel,
  onSubmit,
  submitting,
  disabled = false,
}: BookFormProps) {
  const summaryErrors = (Object.keys(errors) as (keyof BookFormValues)[]).map(
    (field) => ({ id: field, message: `${FIELD_LABELS[field]}: ${errors[field]}` }),
  )

  const toggleTag = (id: number) => {
    onChange({
      ...values,
      tagIds: values.tagIds.includes(id)
        ? values.tagIds.filter((t) => t !== id)
        : [...values.tagIds, id],
    })
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onSubmit()
      }}
      className="flex flex-col gap-5"
    >
      <ErrorSummary errors={summaryErrors} />

      <FormField
        id="titleOverride"
        label={FIELD_LABELS.titleOverride}
        hint="外部カタログの書名が実際と違うときだけ入力する"
        error={errors.titleOverride}
        count={{ current: values.titleOverride.length, max: 255 }}
      >
        <input
          id="titleOverride"
          value={values.titleOverride}
          aria-invalid={errors.titleOverride ? true : undefined}
          aria-describedby={describedBy('titleOverride')}
          onChange={(e) =>
            onChange({ ...values, titleOverride: e.target.value })
          }
          className="h-12 w-full rounded-lg border border-ink-400 px-4 text-base"
        />
      </FormField>

      <FormField
        id="summary"
        label={FIELD_LABELS.summary}
        required
        error={errors.summary}
        count={{ current: values.summary.length, max: 100 }}
      >
        <input
          id="summary"
          value={values.summary}
          aria-invalid={errors.summary ? true : undefined}
          aria-describedby={describedBy('summary')}
          onChange={(e) => onChange({ ...values, summary: e.target.value })}
          className="h-12 w-full rounded-lg border border-ink-400 px-4 text-base"
        />
      </FormField>

      <FormField
        id="comment"
        label={FIELD_LABELS.comment}
        required
        error={errors.comment}
        count={{ current: values.comment.length, max: 5000 }}
      >
        <textarea
          id="comment"
          rows={8}
          value={values.comment}
          aria-invalid={errors.comment ? true : undefined}
          aria-describedby={describedBy('comment')}
          onChange={(e) => onChange({ ...values, comment: e.target.value })}
          className="w-full rounded-lg border border-ink-400 px-4 py-3 text-base leading-relaxed"
        />
      </FormField>

      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm font-bold">
          {FIELD_LABELS.rating} <span className="text-brand-700">（必須）</span>
        </legend>
        <div className="flex gap-2">
          {[1, 2, 3, 4, 5].map((n) => (
            <label
              key={n}
              className="flex h-11 items-center gap-2 rounded-pill border border-ink-400 px-4 text-sm has-[:checked]:border-ink-900 has-[:checked]:bg-ink-900 has-[:checked]:text-white"
            >
              <input
                type="radio"
                name="rating"
                className="sr-only"
                checked={values.rating === n}
                onChange={() => onChange({ ...values, rating: n })}
              />
              {RATING_LABELS[n]}
            </label>
          ))}
        </div>
        {errors.rating && (
          <p role="alert" className="text-[13px] text-brand-800">
            {errors.rating}
          </p>
        )}
      </fieldset>

      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm font-bold">
          {FIELD_LABELS.tagIds}{' '}
          <span className="font-normal text-ink-600">（任意・10個まで）</span>
        </legend>
        <div className="flex flex-wrap gap-2">
          {tags.map((tag) => (
            <label
              key={tag.id}
              className="flex h-11 items-center gap-2 rounded-pill border border-ink-400 px-4 text-sm has-[:checked]:border-ink-900 has-[:checked]:bg-ink-900 has-[:checked]:text-white"
            >
              <input
                type="checkbox"
                className="sr-only"
                checked={tag.id !== undefined && values.tagIds.includes(tag.id)}
                onChange={() => tag.id !== undefined && toggleTag(tag.id)}
              />
              {tag.name}
            </label>
          ))}
        </div>
        {errors.tagIds && (
          <p role="alert" className="text-[13px] text-brand-800">
            {errors.tagIds}
          </p>
        )}
      </fieldset>

      <button
        type="submit"
        disabled={submitting || disabled}
        className="h-12 self-start rounded-lg bg-brand-700 px-8 font-bold text-white hover:bg-brand-800 disabled:cursor-not-allowed disabled:bg-ink-400"
      >
        {submitLabel}
      </button>
    </form>
  )
}
