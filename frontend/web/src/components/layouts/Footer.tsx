export function Footer() {
  return (
    // 一覧にも楽天の書影が出るため、楽天ウェブサービスのクレジットを全画面に置く
    <footer className="border-t border-ink-200 bg-white py-6 text-xs text-ink-600">
      <div className="container-wide flex flex-col gap-1 md:flex-row md:justify-between">
        <p>書誌: openBD ／ 書影: openBD・楽天ブックス</p>
        <a
          href="https://webservice.rakuten.co.jp/"
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand-700 hover:underline"
        >
          Supported by Rakuten Developers
        </a>
      </div>
    </footer>
  )
}
