import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <section className="container-wide py-16 text-center">
      <h1 className="text-xl font-bold">ページが見つかりませんでした</h1>
      <p className="mt-2 text-sm text-ink-600">
        URL が間違っている可能性があります。
      </p>
      <Link
        to="/"
        className="mt-6 inline-block font-bold text-brand-700 hover:underline"
      >
        一覧へ戻る
      </Link>
    </section>
  )
}
