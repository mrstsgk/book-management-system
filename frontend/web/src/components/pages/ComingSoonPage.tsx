import { Link } from 'react-router-dom'

export function ComingSoonPage() {
  return (
    <div className="container-wide bg-ink-100 py-16 text-center">
      <h1 className="text-xl font-bold text-ink-900">準備中</h1>
      <p className="mt-2 text-sm text-ink-600">
        この画面は今後のフェーズで実装します。
      </p>
      <Link
        to="/"
        className="mt-6 inline-block font-bold text-brand-600 hover:underline"
      >
        トップへ戻る
      </Link>
    </div>
  )
}
