import { BookForm } from '@/features/admin-book-editor/components/BookForm'
import { useBookRegistration } from '@/features/admin-book-editor/hooks/useBookRegistration'

export function AdminBookNewRoute() {
  const reg = useBookRegistration()

  return (
    <section className="container-wide flex flex-col gap-6 py-8 md:py-10">
      <h1 className="text-2xl font-extrabold">本を登録</h1>

      <div className="card flex flex-col gap-4 p-6 md:p-7">
        <h2 className="text-lg font-bold">1. ISBN で本を確かめる</h2>
        <div className="flex flex-col gap-1.5">
          <label htmlFor="isbn" className="text-sm font-bold">
            ISBN <span className="text-brand-700">（必須）</span>
          </label>
          <div className="flex gap-2">
            <input
              id="isbn"
              value={reg.isbn}
              onChange={(e) => reg.setIsbn(e.target.value)}
              className="h-12 w-80 rounded-lg border border-ink-400 px-4 text-base"
            />
            <button
              type="button"
              onClick={reg.confirm}
              disabled={reg.isbn.trim() === '' || reg.catalog.isLoading}
              className="h-12 rounded-lg border border-brand-700 px-6 font-bold text-brand-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              確かめる
            </button>
          </div>
        </div>

        {reg.catalog.isLoading && (
          <p role="status" className="text-sm text-ink-600">
            確かめています…
          </p>
        )}
        {reg.catalog.isNotFound && (
          <p role="alert" className="text-sm text-brand-800">
            この ISBN の本は外部カタログに見つかりませんでした。
          </p>
        )}
        {reg.catalog.isError && (
          <p role="alert" className="text-sm text-brand-800">
            確かめられませんでした。時間をおいてもう一度お試しください。
          </p>
        )}
        {reg.catalog.data && (
          <div className="flex gap-5 rounded-lg bg-ink-50 p-4">
            <div className="w-24 flex-shrink-0">
              {reg.catalog.data.coverUrl ? (
                <img
                  src={reg.catalog.data.coverUrl}
                  alt={`${reg.catalog.data.title ?? ''}の書影`}
                  className="w-24 rounded"
                />
              ) : (
                <div className="flex h-[136px] w-24 items-center justify-center rounded bg-ink-200 text-xs text-ink-600">
                  書影なし
                </div>
              )}
            </div>
            <dl className="grid grid-cols-[80px_1fr] gap-y-1.5 text-sm">
              <dt className="text-ink-600">書名</dt>
              <dd className="font-bold">{reg.catalog.data.title}</dd>
              {reg.catalog.data.authors && (
                <>
                  <dt className="text-ink-600">著者</dt>
                  <dd>{reg.catalog.data.authors}</dd>
                </>
              )}
              {reg.catalog.data.publisher && (
                <>
                  <dt className="text-ink-600">出版社</dt>
                  <dd>{reg.catalog.data.publisher}</dd>
                </>
              )}
              {reg.catalog.data.publishedOn && (
                <>
                  <dt className="text-ink-600">出版日</dt>
                  <dd>{reg.catalog.data.publishedOn}</dd>
                </>
              )}
            </dl>
          </div>
        )}
      </div>

      <div className="card flex flex-col gap-4 p-6 md:p-7">
        <h2 className="text-lg font-bold">2. 一言まとめと感想を書く</h2>
        {reg.generalError && (
          <p role="alert" className="text-sm text-brand-800">
            {reg.generalError}
          </p>
        )}
        <BookForm
          values={reg.values}
          onChange={reg.setValues}
          errors={reg.errors}
          tags={reg.tags}
          submitLabel="登録する"
          onSubmit={reg.submit}
          submitting={reg.submitting}
          disabled={!reg.catalog.confirmed}
        />
        {!reg.catalog.confirmed && (
          <p className="text-[13px] text-ink-600">
            先に ISBN を確かめてください。
          </p>
        )}
      </div>
    </section>
  )
}
