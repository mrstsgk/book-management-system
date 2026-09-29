# 書影の取得元の見直し（Google Books → openBD、楽天は撤去）の設計

**日付:** 2026-09-29
**参照:** [ADR](../../adr/2026-09-29-cover-from-google-books-and-openbd.md)、[書影の取得元の調査](#調査の要約)

## 目的

見本の5冊を含め、ほぼすべての本に書影を出す。あわせて、取得元ごとの特別な仕組み（楽天の保持期限・取り直し・削除指示）を無くし、「書影は提供元の URL を表示するだけ」という1つの考え方で説明できる形にする。

## 決定

書影は次の順で、最初に見つかったものを使う。どれにも無ければ書影なし（画面はプレースホルダー）。

| 順 | 取得元 | 条件 | 画面に必要な表示 |
|---|---|---|---|
| 1 | Google Books API | `GOOGLE_BOOKS_API_KEY` を設定したときだけ問い合わせる | 「Powered by Google」と、その本の Google Books のページへのリンク |
| 2 | openBD | 常に（書誌を取るついでに返ってくる） | なし（openBD の書影は、版元ドットコムで出版社が「利用可」にしたものだけ） |
| 3 | なし | — | 「書影なし」の枠 |

- 書誌（書名・著者・出版社・出版日）は今までどおり openBD だけから取る。Google Books の書誌は使わない（書誌の出どころを1つに保つため）
- 楽天ブックスは撤去する。楽天の書影・商品ページ・取得日時・削除指示の有無、`DELETE /api/books/{id}/rakuten`、起動時と1日1回の取り直し、楽天のクレジット表示をすべて消す。未マージの #88 は閉じる
- 「版元ドットコムで利用可の書影」は openBD の書影と同じものなので、版元ドットコムへは別に問い合わせない

## 仕組み（理解しやすさのための形）

```text
UseCase ──> domain/book.BookCatalog（Lookup: 書誌＋書影）
                 ▲
infrastructure/gateway/catalog.New(openbd, googlebooks)
   1. openbd.Lookup        → 書誌と openBD の書影
   2. googlebooks.FindCover → 見つかれば書影を Google Books のものに差し替える
      （キー未設定なら googlebooks は nil で、2 を飛ばす。失敗は warn ログだけで 1 の結果を返す）
```

- UseCase から見える外部カタログのポートは今までどおり `BookCatalog` 1つ。取得元の順番は `gateway/catalog` の1関数だけに書く（どこを読めば順番が分かるかを1か所にする）
- `gateway/googlebooks` は書影だけを返す小さな Gateway。`volumes?q=isbn:<ISBN>&key=<キー>` の先頭の結果の `imageLinks.thumbnail`（`http` なら `https` に直す）と `volumeInfo.infoLink`（Google Books のページ）を使う。`imageLinks` が無ければ「書影なし」
- Domain の `Cover` VO は `url`・`source`（`openbd` / `googlebooks`）・`pageURL`（提供元の本のページ。Google Books では必須、openBD では空）だけを持つ。期限や取得日時は持たない

## 既存の本の書影

- 登録と更新のたびに、上の順で書影を取り直す（今までどおり）
- API の起動時に1回だけ、**書影の無い本**を上の順で問い合わせて埋める（キーを後から設定しても、見本の本に書影が付くように）。書影が変わらなければ保存しない。定期実行はしない（書影に期限が無くなったため）

## API

- `coverSource`: `openbd` / `googlebooks`
- `coverProductUrl` を `coverPageUrl`（提供元の本のページ。Google Books の書影のときだけ）に置き換える
- `rakutenDisabled` と `DELETE /api/books/{id}/rakuten` を消す

## 画面

- 詳細: Google Books の書影のとき、書影の下に「Powered by Google」と「Google Books で見る」リンク（`coverPageUrl`）を出す。楽天の表示は消す
- 一覧: Google Books の書影のカードには、カードの外（書影の下）に小さく「Google Books」リンクを出す（リンクの入れ子にしない）
- フッター: 「書誌: openBD ／ 書影: Google Books・openBD」。楽天のクレジットは消す
- 画面イメージ（`docs/design/public-screens/`）も同じように直す

## DB

- `000007`: 楽天の書影を持つ行の書影を消し、`ck_book_rakuten_cover` と `cover_product_url`・`cover_fetched_at`・`rakuten_disabled` を削除する。`cover_source` の CHECK は `('openbd')` にする
- `000008`: `cover_page_url` を追加し、`cover_source` の CHECK を `('openbd', 'googlebooks')` にする。`cover_source = 'googlebooks'` なら `cover_page_url` 必須、それ以外は NULL
- `docs/db/backend-schema.{json,md}` を同じ PR で直す

## 調査の要約

| 取得元 | 5冊で取れた数 | アプリで表示 | 備考 |
|---|---|---|---|
| Google Books | 4/5 | 可（Google のロゴと本ごとのリンクが条件） | 無料の API キーが前提（キーなしは共有の日次枠で 429） |
| 版元ドットコム（＝openBD の書影） | 利用可のものだけ | 利用可の本だけ可 | 見本の4冊は「利用可否不明」 |
| 国立国会図書館 | — | 不可 | 2026-03-31 に外部提供を終了 |
| 楽天ブックス | — | 可 | 3か月の保持期限と削除指示への対応が要る（今回撤去） |

## やらないこと

- 書影の画像そのものを保存すること（どの取得元も、そのまま表示する以外を認めていない）
- README に書影の写ったスクショを載せること（別途、書影をぼかして撮る）
