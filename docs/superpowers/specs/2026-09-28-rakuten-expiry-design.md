# 楽天由来の情報の期限管理 設計

**日付:** 2026-09-28
**対象:** 要件定義 [§1.2 楽天由来の情報の持ち方](../../specifications.md)、および §1.1「楽天の商品ページ」・§2.2 の楽天の商品ページへのリンク
**変更前の形:** [読んだ本の API 設計](./2026-09-28-reading-api-design.md)、ADR [楽天は書影を補う用途だけに使う](../../adr/2026-09-28-rakuten-for-cover-only.md)。この文書はそこからの差分だけを書く

## 目的

- 楽天ウェブサービスの規約（価格・販売可否以外の情報は最長3か月まで、削除の指示にはすぐ従う）を守ったうえで、楽天の書影を使い続ける
- 楽天の書影を見せるときに必要な、楽天の商品ページへのリンクを返す（今は取得も保存もしていない）

## やらないこと

- 画面側のクレジット・注記の表示（単位2・3）。API は必要な値を返すだけ
- openBD 由来の書影の期限管理（openBD には保持期限が無い）
- 取り直しの間隔・期限の設定値化。規約の値は固定なので定数にする

## 1. 期限の決めごと

| 定数 | 値 | 理由 |
|---|---|---|
| 保持期限 | 取得から89日 | 連続する3か月で最も短いのは89日（平年の2〜4月）。89日で切れば、どの日に取得しても規約の「最長3か月」を超えない |
| 取り直しの開始 | 期限の7日前（取得から82日） | 要件定義 §1.2「残り7日以内」 |

暦の月で足す（`AddDate(0, 3, 0)`）方式は採らない。11/30 に3か月を足すと 2/30 が 3/2 に正規化されるように、月末の取得で期限が3か月を超えてしまうため。固定の日数なら SQL と Go で同じ規則を書けることも理由。

## 2. ドメイン

楽天の書影を「画像 URL・商品ページ URL・取得日時」の3つ組でしか作れないようにする（商品ページや取得日時の無い楽天の書影という、不正な状態を作れなくする）。

```go
func NewCover(rawURL string, source CoverSource) (Cover, error)  // openBD 専用にし、楽天を渡すとエラー
func NewRakutenCover(imageURL, productURL string, fetchedAt time.Time) (Cover, error)

func (c Cover) ProductURL() string              // 楽天の商品ページ。openBD なら空
func (c Cover) FetchedAt() time.Time            // 楽天の取得日時。openBD ならゼロ値
func (c Cover) IsExpired(now time.Time) bool    // 楽天で取得から89日以上。openBD は常に false
func (c Cover) NeedsRefresh(now time.Time) bool // 楽天で取得から82日以上
```

`Book` に次を足す。

| 追加 | 内容 |
|---|---|
| `RakutenDisabled bool` | 楽天から削除の指示を受けた本。以後、この本には楽天の書影を付けない（取り直しで元に戻らないようにするため） |
| `DropExpiredCover(now)` | 楽天の書影が期限切れなら外す |
| `DisableRakuten()` | 楽天の書影を外し、`RakutenDisabled` を立てる |
| `RefreshCatalog` の変更 | `RakutenDisabled` の本に楽天の書影が来たら、書影なしとして扱う |

時刻は呼び出し側（ユースケース）が `now` として渡す。ドメインは時計を持たない（テストで時刻を固定できるようにするため）。

## 3. テーブル（`000005`）

```sql
ALTER TABLE book ADD COLUMN cover_product_url VARCHAR(2048);
ALTER TABLE book ADD COLUMN cover_fetched_at TIMESTAMPTZ;
ALTER TABLE book ADD COLUMN rakuten_disabled BOOLEAN NOT NULL DEFAULT FALSE;

-- 既存の楽天の書影は商品ページも取得日時も分からないので外す（本を更新すれば取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'rakuten';

ALTER TABLE book ADD CONSTRAINT ck_book_rakuten_cover CHECK (
    (cover_source = 'rakuten' AND cover_product_url IS NOT NULL AND cover_fetched_at IS NOT NULL)
    OR (cover_source IS DISTINCT FROM 'rakuten' AND cover_product_url IS NULL AND cover_fetched_at IS NULL)
);
```

既存の楽天の書影を外すのは、このアプリがローカルでしか動かさない（要求定義 §4 デプロイしない）ため。本番のデータを守る仕組みは要らない。

## 4. 外部カタログ

楽天のゲートウェイが、検索結果の `itemUrl`（商品ページ）を読み、`NewRakutenCover(largeImageUrl, itemUrl, now)` で書影を作る。`itemUrl` が空なら楽天の書影は付けない（商品ページへリンクできない書影は規約上使えないため）。

## 5. 期限切れを返さない（参照側）

`book.Query` の詳細・一覧で、楽天の書影が期限切れ（`cover_fetched_at <= NOW() - INTERVAL '89 days'`）なら書影を返さない（`coverUrl`・`coverSource`・`coverProductUrl` を null）。取り直し・消去が走る前でも、期限切れの値を画面に出さないための保険。

`Response`・`ListItemResponse` に `coverProductUrl`（nullable）を足す。

## 6. 取り直しと消去

### 6.1 ユースケース `RefreshRakutenCoversUsecase`（`usecase/book/command`）

1. 取得から82日以上たった楽天の書影を持つ本を Repository から読む（`book.Repository.FindRakutenRefreshTargets(ctx, fetchedBefore time.Time)`）
2. 1冊ずつ外部カタログに問い合わせ、書影を取り直して保存する
3. 取り直せず（障害・該当なし）、期限も切れていたら、書影を外して保存する。期限前ならそのままにして次の契機を待つ
4. 保存が `ErrConflict`（その間に自分で更新した）なら飛ばす。その更新で書影は取り直されている
5. 1冊の失敗で残りを止めない。失敗は warn ログに残す

### 6.2 契機

| 契機 | どう呼ぶか |
|---|---|
| API の起動時 | `run()` で HTTP サーバを立てた後、別の goroutine で1回呼ぶ（外部カタログが遅くても起動を待たせない） |
| 稼働中は1日1回 | 同じ goroutine で24時間ごとに呼ぶ。シャットダウン時の context のキャンセルで止める |
| 本の更新時 | 既存の `UpdateUsecase` が毎回書影を取り直している。取り直せなかったときに `DropExpiredCover(now)` を足す |

登録時（`RegisterUsecase`）は取得した直後なので、取得日時が今の書影が付くだけ。

## 7. 楽天からの削除指示への対応（`DELETE /api/books/{id}/rakuten`）

管理画面から、対象の本の楽天由来の情報を消す。認証必要。

- `DisableRakuten()` して保存する。以後その本には楽天の書影を付けない
- 楽天の書影が付いていない本でも 204（「楽天の情報を持たない状態」にするという操作なので、既にそうなら成功）
- 存在しない本は 404
- `version` は受け取らない（削除の指示には、画面で最後に読んだ内容に関係なく従う必要があるため）。内部では読んだ直後の version で保存し、競合したら1回だけ読み直してやり直す

`Response` に `rakutenDisabled` を足す（管理画面で「楽天の情報を消した本」だと分かるように）。

## 8. エラー

| 状況 | 結果 |
|---|---|
| 楽天の情報を消す本が存在しない | 404 |
| 取り直しの失敗（起動時・定期） | ログだけ。API の応答には影響しない |

## 9. テスト

| 対象 | 確かめること |
|---|---|
| `Cover` | 楽天は商品ページ・取得日時が必須、openBD に楽天を渡すとエラー、`IsExpired` の境界（88日台・89日）、`NeedsRefresh` の境界（81日台・82日）、openBD は期限切れにならない |
| `Book` | `DropExpiredCover`、`DisableRakuten` の後は `RefreshCatalog` で楽天の書影が付かない（openBD の書影は付く） |
| 楽天のゲートウェイ | `itemUrl` を読む、`itemUrl` が空なら書影なし |
| `book.Repository`（契約テスト） | 商品ページ・取得日時・`rakuten_disabled` を保存して読める、`FindCoverRefreshTargets` の境界 |
| `book.Query`（契約テスト） | 期限切れの楽天の書影を返さない（詳細・一覧）、期限前は返す、openBD の書影は古くても返す |
| `RefreshRakutenCoversUsecase`（Fake） | 取り直せたら更新、取り直せず期限切れなら外す、期限前なら触らない、`ErrConflict` を飛ばす、1冊の失敗で止まらない |
| `UpdateUsecase`（Fake） | 取り直せず期限切れなら外して保存する |
| Handler | `DELETE /api/books/{id}/rakuten` の認証要否・204・404 |
| 起動（実 DB） | 起動しても取り直しを待たずにリクエストを受け付ける |

## 10. PR の分け方

1. 楽天由来の情報を持つ（§2〜§5。`Cover` の変更・マイグレーション・ゲートウェイの `itemUrl`・参照側で期限切れを返さない・`coverProductUrl`）
2. 取り直しと消去（§6。ユースケース・起動時と1日1回の goroutine・更新時の消去）
3. 楽天からの削除指示への対応（§7。`RakutenDisabled`・`DELETE` API）

1 の後に 2・3 を並行して進められる。1 は[見本データ投入](./2026-09-28-sample-data-design.md)（起動時に書影を取る）・[一覧の検索](./2026-09-28-book-list-search-design.md)（`book.Query` を変える）と触る箇所が重なるので、それらの後にマージする側が合わせる。

## 追記（2026-09-29）: 書影なしの本も取り直しの対象にする

楽天のキーを設定せずに見本データを入れると、見本の本は書影なしで保存される。取り直しの対象が「楽天の書影を持ち期限が近い本」だけだったため、後からキーを設定しても書影が付かず、[要件定義 §1.2](../../specifications.md#12-楽天由来の情報の持ち方)の「見本データの書影はキーを設定して起動したときに楽天から取得する」を満たしていなかった。

- 起動時と1日1回の取り直しの対象に、書影なしの本（`cover_source IS NULL`）を加える。楽天から削除の指示を受けた本（`rakuten_disabled`）は含めない
- 取り直し対象を探す `book.Repository` のメソッドは `FindRakutenRefreshTargets` から `FindCoverRefreshTargets` に改めた（楽天に限らなくなったため）
- 書影なしの本で書影が見つからない・取れないときは保存しない（`version` を進めず、同時に編集している画面と無駄に競合させないため）
- openBD の書影を持つ本は対象にしない（openBD の書影に保持期限は無い）
