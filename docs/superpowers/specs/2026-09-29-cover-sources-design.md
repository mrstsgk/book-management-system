# 書影の取得元の見直し（openBD だけ、楽天は撤去）の設計

**日付:** 2026-09-29（2026-09-30 改訂）
**参照:** [ADR（楽天の撤去）](../../adr/2026-09-29-cover-from-google-books-and-openbd.md)、[ADR（openBD だけにする）](../../adr/2026-09-30-cover-from-openbd-only.md)、[書影の取得元の調査](#調査の要約)

> **変更（2026-09-30）:** 当初は「Google Books → openBD → なし」の順にする設計だったが、Google Books も API キーが必要で、キーを持たない人の環境では書影が出ないため使わないことにした（[ADR](../../adr/2026-09-30-cover-from-openbd-only.md)）。Google Books の配線（#93）は閉じ、Gateway（#89）は消した。以下は改訂後の設計。

## 目的

取得元ごとの特別な仕組み（楽天の保持期限・取り直し・削除指示）を無くし、「書影は提供元の URL を表示するだけ」という1つの考え方で説明できる形にする。キーを持たない人の環境でも、同じ書影が見えるようにする。

## 決定

| 取得元 | 条件 | 画面に必要な表示 |
|---|---|---|
| openBD | 常に（書誌を取るついでに返ってくる） | なし（openBD の書影は、版元ドットコムで出版社が「利用可」にしたものだけ） |
| なし | openBD に書影が無い本 | 「書影なし」の枠 |

- 書誌（書名・著者・出版社・出版日）も書影も openBD だけから取る
- 楽天ブックスは撤去する。楽天の書影・商品ページ・取得日時・削除指示の有無、`DELETE /api/books/{id}/rakuten`、起動時と1日1回の取り直し、楽天のクレジット表示をすべて消す（#90）
- 「版元ドットコムで利用可の書影」は openBD の書影と同じものなので、版元ドットコムへは別に問い合わせない
- 書影を出したい本は、openBD に書影がある本を選ぶ

## 仕組み

```text
UseCase ──> domain/book.BookCatalog（Lookup: 書誌＋書影）
                 ▲
infrastructure/gateway/openbd（書誌と openBD の書影）
```

- UseCase から見える外部カタログのポートは `BookCatalog` 1つで、実装は openBD だけ（`cmd/api` の `newCatalog`）
- Domain の `Cover` VO は `url` と `source`（`openbd`）だけを持つ。期限や取得日時は持たない

## 既存の本の書影

- 登録と更新のたびに、openBD から書影を取り直す（今までどおり）
- 画像は保存せず提供元の URL を都度表示するだけなので、openBD が書誌・書影の削除を求めても、次の登録・更新で取り直した結果にそのまま反映される。削除要請を受けて何かを消す専用の仕組みは持たない

## API

- `coverSource`: `openbd`
- `coverProductUrl`・`rakutenDisabled` と `DELETE /api/books/{id}/rakuten` を消す

## 画面

- 書影があればその画像、無ければ「書影なし」の枠。楽天の表示は消す
- フッター: 「書誌・書影: openBD」

## DB

- `000007`: 楽天の書影を持つ行の書影を消し、`ck_book_rakuten_cover` と `cover_product_url`・`cover_fetched_at`・`rakuten_disabled` を削除する。`cover_source` の CHECK は `('openbd')` にする
- `docs/db/backend-schema.{json,md}` を同じ PR で直す

## 調査の要約

| 取得元 | 5冊で取れた数 | アプリで表示 | 備考 |
|---|---|---|---|
| Google Books | 4/5 | 可（Google のロゴと本ごとのリンクが条件） | 無料の API キーが前提（キーなしは共有の日次枠で 429） |
| 版元ドットコム（＝openBD の書影） | 利用可のものだけ | 利用可の本だけ可 | 見本の4冊は「利用可否不明」 |
| 国立国会図書館 | — | 不可 | 2026-03-31 に外部提供を終了 |
| 楽天ブックス | — | 可 | 3か月の保持期限と削除指示への対応が要る（撤去） |

Google Books は API キーが要るため、2026-09-30 に使わないことにした（上の「変更」を参照）。

## やらないこと

- 書影の画像そのものを保存すること（どの取得元も、そのまま表示する以外を認めていない）
- README に画面のスクショを載せること
