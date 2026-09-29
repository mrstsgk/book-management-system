import { Link } from 'react-router-dom'
import type { BookResponse } from '@/api/generated/api.schemas'
import { RatingStars } from '@/components/ui/RatingStars'

type BookDetailViewProps = {
  book: BookResponse
}

const externalLinkClass =
  'inline-flex h-11 items-center rounded-lg border border-brand-700 px-5 text-[15px] font-bold text-brand-700 no-underline hover:bg-brand-50'

function BookCover({ book }: BookDetailViewProps) {
  return (
    <div className="flex w-[220px] shrink-0 flex-col gap-2.5">
      {book.coverUrl ? (
        <img
          src={book.coverUrl}
          alt={`${book.title ?? ''}の書影`}
          className="w-[220px] rounded object-contain"
        />
      ) : (
        <div className="flex h-[312px] w-[220px] items-center justify-center rounded bg-ink-200 text-[13px] text-ink-600">
          書影なし
        </div>
      )}
      {/* Google Books の規約では、書影と一緒に「Powered by Google」とその本の Google Books へのリンクが要る */}
      {book.coverSource === 'googlebooks' && book.coverPageUrl && (
        <p className="flex flex-col gap-1 text-xs text-ink-600">
          <span>Powered by Google</span>
          <a
            href={book.coverPageUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-brand-700 hover:underline"
          >
            Google Books で見る <span aria-hidden="true">↗</span>
          </a>
        </p>
      )}
    </div>
  )
}

function ExternalLinks({ book }: BookDetailViewProps) {
  if (!book.amazonUrl) return null
  return (
    <div className="mt-2 flex flex-wrap gap-3">
      <a
        href={book.amazonUrl}
        target="_blank"
        rel="noopener noreferrer"
        className={externalLinkClass}
      >
        Amazonで見る <span aria-hidden="true">↗</span>
      </a>
    </div>
  )
}

export function BookDetailView({ book }: BookDetailViewProps) {
  const facts = [
    ['著者', book.authors],
    ['出版社', book.publisher],
    ['出版日', book.publishedOn],
    ['ISBN', book.isbn],
  ].filter((f): f is [string, string] => Boolean(f[1]))

  return (
    <div className="flex flex-col gap-6">
      <Link to="/" className="w-fit text-sm text-brand-700 hover:underline">
        ← 一覧へ戻る
      </Link>

      <article className="card flex flex-col gap-9 p-6 md:p-10">
        <div className="flex flex-col gap-8 md:flex-row md:gap-10">
          <BookCover book={book} />

          <div className="flex min-w-0 flex-grow flex-col gap-4">
            {book.tags && book.tags.length > 0 && (
              <ul className="flex flex-wrap gap-2">
                {book.tags.map((tag) => (
                  <li
                    key={tag}
                    className="rounded-pill bg-ink-100 px-3 py-1 text-[13px] text-ink-600"
                  >
                    {tag}
                  </li>
                ))}
              </ul>
            )}
            <h1 className="break-words text-2xl font-extrabold leading-snug md:text-[30px]">
              {book.title}
            </h1>
            {book.rating !== undefined && (
              <RatingStars rating={book.rating} size="lg" />
            )}
            <dl className="grid grid-cols-[96px_minmax(0,1fr)] gap-y-2 text-[15px]">
              {facts.map(([label, value]) => (
                <div key={label} className="contents">
                  <dt className="text-ink-600">{label}</dt>
                  <dd className="break-words">{value}</dd>
                </div>
              ))}
            </dl>
            <ExternalLinks book={book} />
          </div>
        </div>

        <section className="flex flex-col gap-2.5">
          <h2 className="text-lg font-bold">一言まとめ</h2>
          <p className="leading-loose">{book.summary}</p>
        </section>
        <section className="flex flex-col gap-2.5">
          <h2 className="text-lg font-bold">感想</h2>
          <p className="max-w-[760px] whitespace-pre-wrap break-words leading-loose">
            {book.comment}
          </p>
        </section>
      </article>
    </div>
  )
}
