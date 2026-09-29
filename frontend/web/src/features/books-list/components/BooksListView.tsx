import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import type {
  BookListItemResponse,
  TagResponse,
} from '@/api/generated/api.schemas'
import { EmptyState } from '@/components/states/EmptyState'
import { ErrorState } from '@/components/states/ErrorState'
import { LoadingState } from '@/components/states/LoadingState'
import { RatingStars } from '@/components/ui/RatingStars'

type BooksListViewProps = {
  books: BookListItemResponse[]
  total: number
  query?: string
  tagId?: number
  tags?: TagResponse[]
  hasFilter: boolean
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  hasMore: boolean
  isLoadingMore: boolean
  onLoadMore: () => void
  onSearch: (query: string) => void
  onSelectTag: (tagId: number | undefined) => void
  onClear: () => void
}

export function BooksListView(props: BooksListViewProps) {
  return (
    <section className="container-wide flex flex-col gap-6 py-8 md:py-10">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-extrabold md:text-[28px]">読んだ本</h1>
        {!props.isLoading && !props.isError && (
          <p className="text-sm text-ink-600">{props.total}冊</p>
        )}
      </div>
      <SearchPanel {...props} />
      <BooksListBody {...props} />
    </section>
  )
}

function SearchPanel({
  query,
  tagId,
  tags,
  onSearch,
  onSelectTag,
}: BooksListViewProps) {
  const submit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    onSearch(String(new FormData(e.currentTarget).get('q') ?? ''))
  }

  return (
    <div className="card flex flex-col gap-4 px-4 py-4 md:px-6 md:py-5">
      {/* key で URL の q が外から変わったとき（条件を外したとき等）に入力欄を作り直す */}
      <form
        key={query ?? ''}
        role="search"
        onSubmit={submit}
        className="flex flex-col gap-1.5"
      >
        <label htmlFor="books-q" className="text-sm font-bold">
          キーワード（書名・著者）
        </label>
        <div className="flex gap-2">
          <input
            id="books-q"
            name="q"
            type="search"
            defaultValue={query}
            placeholder="例: 設計、Kleppmann"
            className="h-12 min-w-0 flex-1 rounded-lg border border-ink-500 px-4 text-base"
          />
          <button
            type="submit"
            className="h-12 rounded-lg bg-brand-700 px-5 font-bold text-white hover:bg-brand-800 md:px-7"
          >
            検索
          </button>
        </div>
      </form>
      {tags && (
        <div
          role="group"
          aria-labelledby="books-tags-label"
          className="flex flex-col gap-2"
        >
          <p id="books-tags-label" className="text-sm font-bold">
            分野タグ
          </p>
          <div className="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1 md:flex-wrap md:overflow-visible">
            <TagButton
              label="すべて"
              pressed={tagId === undefined}
              onClick={() => onSelectTag(undefined)}
            />
            {tags.map((tag) => (
              <TagButton
                key={tag.id}
                label={tag.name ?? ''}
                pressed={tag.id === tagId}
                onClick={() => onSelectTag(tag.id)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function TagButton({
  label,
  pressed,
  onClick,
}: {
  label: string
  pressed: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      aria-pressed={pressed}
      onClick={onClick}
      className={`h-11 shrink-0 rounded-pill border px-4 text-sm ${
        pressed
          ? 'border-ink-900 bg-ink-900 text-white'
          : 'border-ink-400 bg-white text-ink-900 hover:border-ink-700'
      }`}
    >
      {label}
    </button>
  )
}

function BooksListBody({
  books,
  total,
  hasFilter,
  isLoading,
  isError,
  onRetry,
  hasMore,
  isLoadingMore,
  onLoadMore,
  onClear,
}: BooksListViewProps) {
  if (isLoading) return <LoadingState />
  if (isError) {
    return (
      <ErrorState
        title="本の一覧を読み込めませんでした"
        description="時間をおいてもう一度お試しください。"
        onRetry={onRetry}
      />
    )
  }
  if (books.length === 0) {
    return hasFilter ? (
      <EmptyState
        title="条件に当てはまる本はありません"
        description="キーワードを変えるか、分野タグを「すべて」に戻してください。"
        action={
          <button
            type="button"
            onClick={onClear}
            className="h-11 rounded-lg border border-brand-700 bg-white px-6 font-bold text-brand-700 hover:bg-brand-50"
          >
            検索と絞り込みを外す
          </button>
        }
      />
    ) : (
      <EmptyState title="まだ本が登録されていません" />
    )
  }

  return (
    <>
      {/* Google Books の規約では、その書影を出す結果の近くに「Powered by Google」が要る */}
      {books.some(isGoogleBooksCover) && (
        <p className="text-right text-xs text-ink-600">Powered by Google</p>
      )}
      <ul className="grid grid-cols-1 gap-3 md:grid-cols-2 md:gap-5">
        {books.map((book) => (
          <li key={book.id} className="flex flex-col gap-1">
            <BookCard book={book} />
            {/* カードのリンクの中に入れるとリンクの入れ子になるので、カードの外に置く */}
            {isGoogleBooksCover(book) && (
              <a
                href={book.coverPageUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="w-fit px-1 text-xs text-brand-700 hover:underline"
              >
                Google Books <span aria-hidden="true">↗</span>
              </a>
            )}
          </li>
        ))}
      </ul>
      <div className="flex flex-col items-center gap-2 pt-2">
        <p className="text-[13px] text-ink-600">
          {total}冊中 {books.length}冊を表示
        </p>
        {isLoadingMore && (
          <p role="status" className="text-sm">
            続きを読み込み中…
          </p>
        )}
        {hasMore && (
          <button
            type="button"
            onClick={onLoadMore}
            disabled={isLoadingMore}
            className="h-12 w-full rounded-lg border border-brand-700 bg-white px-10 font-bold text-brand-700 hover:bg-brand-50 disabled:border-ink-400 disabled:bg-ink-100 disabled:text-ink-500 md:w-auto"
          >
            もっと見る
          </button>
        )}
      </div>
    </>
  )
}

function isGoogleBooksCover(book: BookListItemResponse): boolean {
  return book.coverSource === 'googlebooks' && Boolean(book.coverPageUrl)
}

function BookCard({ book }: { book: BookListItemResponse }) {
  return (
    <Link
      to={`/books/${book.id}`}
      className="card flex gap-3.5 p-3.5 text-ink-900 hover:shadow-soft md:h-44 md:gap-5 md:p-5"
    >
      {book.coverUrl ? (
        // 書名はリンク内の見出しで読み上げるため、書影は飾りとして扱う
        <img
          src={book.coverUrl}
          alt=""
          className="h-[102px] w-[72px] shrink-0 rounded object-cover md:h-[136px] md:w-24"
        />
      ) : (
        <div className="flex h-[102px] w-[72px] shrink-0 items-center justify-center rounded bg-ink-200 text-[11px] text-ink-600 md:h-[136px] md:w-24 md:text-xs">
          書影なし
        </div>
      )}
      <div className="flex min-w-0 flex-col gap-1 md:gap-1.5">
        <h2 className="line-clamp-2 break-words text-[15px] font-bold leading-snug md:text-[17px]">
          {book.title}
        </h2>
        <p className="truncate text-xs text-ink-600 md:text-[13px]">
          {book.authors}
        </p>
        <p className="line-clamp-2 break-words text-[13px] leading-relaxed md:text-sm">
          {book.summary}
        </p>
        <div className="mt-auto flex flex-wrap items-center gap-2 md:gap-3">
          {book.rating !== undefined && <RatingStars rating={book.rating} />}
          {book.tags?.map((tag) => (
            <span
              key={tag}
              className="rounded-pill bg-ink-100 px-2.5 py-0.5 text-[11px] text-ink-600 md:text-xs"
            >
              {tag}
            </span>
          ))}
        </div>
      </div>
    </Link>
  )
}
