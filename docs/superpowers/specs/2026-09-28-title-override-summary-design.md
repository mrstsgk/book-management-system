# 書名の上書き・一言まとめ 設計

**日付:** 2026-09-28
**対象:** バックエンド拡張 b（[要件定義](../../specifications.md) §1.1 の「書名の上書き」「一言まとめ」）
**変更前の形:** [読んだ本の API 設計（既存）](./2026-09-28-reading-api-design.md)。この文書はそこからの差分だけを書く

## 目的

- 外部カタログの書名が実際の書名と違う本（例: openBD の「AWS認定ソリューションアーキテクト-アソシエイト教科書」に「徹底攻略」「第3版」が付かない）を、自分で正しい書名にできるようにする
- 一覧で見せる短い要約（一言まとめ）を本ごとに持つ

## やらないこと

- 書誌の列を NULL にすること、表示する書名で ISBN を代わりに見せること（楽天の規約対応 e で扱う）
- 画面（公開画面・管理画面の単位で扱う）

## 1. テーブル

マイグレーション `000002` で `book` に列を足す。

| 列 | 型 | 意味 |
|---|---|---|
| `title_override` | `VARCHAR(255) NULL` | 自分で上書きした書名。NULL なら上書きしていない |
| `summary` | `VARCHAR(100) NOT NULL` | 一言まとめ |

- `summary` は既存の行があっても列を足せるよう、既定値 `''` 付きで足してから既定値を外す。空文字の行はドメインの検証で読み込めないため、既存の行がある環境では手で埋めるか作り直す（本番環境は無い）
- `COMMENT ON COLUMN` を付け、`docs/db/backend-schema.{json,md}` も更新する

## 2. ドメイン（`domain/book`）

### 2.1 `Title`（新しい VO）

- 1〜255 文字。前後の空白を除く。タブや改行などの制御文字は受け付けない（トリム前に検査する）
- 今の `Bibliography` の書名と同じ規則なので、`Bibliography` の書名もこの `Title` で持つように揃える（規則を1か所にする）。`Bibliography.Title()` は今どおり文字列を返す

### 2.2 `Summary`（新しい VO）

- 1〜100 文字。前後の空白を除く。一覧に出す1行の文なので、改行を含む制御文字は受け付けない

### 2.3 `Book`

| 追加・変更 | 内容 |
|---|---|
| フィールド | `TitleOverride *Title`（nil なら上書きなし）、`Summary Summary` |
| `New` | `New(isbn, bibliography, cover, summary, comment, rating)` に `summary` を足す。上書きは `OverrideTitle` で付ける |
| `ChangeReview` | `ChangeReview(summary, comment, rating, version)`。一言まとめも感想と同じく自分が書く紹介文なので、同じ操作で差し替える |
| `OverrideTitle` | `OverrideTitle(title *Title)`。nil で上書きを外す。呼び出し側の変数を後から変えても本が変わらないようコピーして持つ |
| `RefreshCatalog` | 変更なし。外部カタログを取り直しても `TitleOverride` には触れない |

### 2.4 Read Model

| Read Model | 追加・変更 |
|---|---|
| `BookDetail` | `Title` を「表示する書名」にする。`CatalogTitle string`（外部カタログの書名）、`TitleOverride *string`、`Summary string` を足す |
| `BookListItem` | `Title` を「表示する書名」にする。`Summary string` を足す |

表示する書名 = 上書きがあればそれ、無ければ外部カタログの書名。管理画面の編集で上書きと元の書名の両方を見せるため、詳細では分けても返す。

## 3. 永続化（`infrastructure/postgres/book`）

- Repository: `title_override`・`summary` を読み書きする。読み込み時は VO で検証し直す（今の `adapt` と同じ）
- Query: 表示する書名を SQL の `COALESCE(title_override, title)` で決める。Read Model には結果の文字列だけを載せる

## 4. ユースケース（`usecase/book/command`）

| Command | 追加 |
|---|---|
| `RegisterCommand` | `Summary string`、`TitleOverride string`（空文字なら上書きなし） |
| `UpdateCommand` | 同上。空文字なら上書きを外す |

- どちらも入力の検証（ISBN・一言まとめ・感想・評価・書名の上書き）を外部カタログへの問い合わせより先に済ませる（今の方針どおり）

## 5. HTTP / OpenAPI（`presentation/http/book`）

| エンドポイント | 追加する入力 |
|---|---|
| `POST /api/books` | `summary`（必須、`max=100`）、`titleOverride`（任意、`max=255`） |
| `PUT /api/books/{id}` | 同上。PUT は全体の置き換えなので、`titleOverride` を省略するか空にすると上書きを外す |

| レスポンス | 追加 |
|---|---|
| `BookResponse` | `summary`、`catalogTitle`、`titleOverride`（null 可）。`title` は表示する書名 |
| `BookListItemResponse` | `summary`。`title` は表示する書名 |

swag で OpenAPI を排出し直し、フロントの型を再生成する。

## 6. エラー

| 状況 | 結果 |
|---|---|
| `summary` が空・100 文字超・改行を含む | 400 |
| `titleOverride` が 255 文字超・制御文字を含む | 400 |
| それ以外 | 今と同じ（カタログに無い ISBN は 400、重複は 409、版の不一致は 409） |

## 7. テスト

| 対象 | 確かめること |
|---|---|
| `Title` / `Summary` | 上限ちょうど・上限+1、下限（空・空白だけ）、前後の空白の除去（除去前と除去後の両方）、制御文字 |
| `Bibliography` | 書名の規則が `Title` と同じであること（既存テストが通ること） |
| `Book` | `OverrideTitle` で付ける・nil で外す・コピーして持つ。`RefreshCatalog` 後も上書きが残る。`ChangeReview` で一言まとめが変わる |
| Repository（契約テスト） | 上書き・一言まとめの保存と読み込み、上書きを外すと NULL になる |
| Query（契約テスト） | 表示する書名が上書き優先で、無ければカタログの書名。`CatalogTitle` と `TitleOverride` が別に取れる。Read Model が VO 型でなく `string` であること |
| ユースケース | 一言まとめ・書名の上書きが不正なら外部カタログも Repository も呼ばれない。更新で上書きを外せる |
| Handler | 必須の欠落と上限超えが 400 とフィールドのエラーになる。`titleOverride` 省略時に上書きなしで渡る |

## 8. PR の分け方

1. ドメイン（`Title`・`Summary`・`Book` の変更と Read Model）とこの設計書
2. それ以外（マイグレーション・DB の資料・Repository/Query・ユースケース・Handler・OpenAPI・フロントの型の再生成）
