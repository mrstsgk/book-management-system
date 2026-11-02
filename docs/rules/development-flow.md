# 開発の流れ

要求からリリースまでを 1 本につないだ流れ。各段階の細かい規則は、表の「正」の文書に従う。本書は段階のつながり（何ができたら次へ進むか）だけを定める。

```mermaid
flowchart LR
  R["1. 要求・要件<br/>docs/requirements.md<br/>docs/specifications.md"] --> D["2. 設計<br/>docs/design/（概要設計・モック）<br/>docs/superpowers/specs/"]
  D -->|設計の合意| DD["3. 詳細設計<br/>openapi/ の YAML"]
  DD -->|設計レビュー OK・生成| I["4. 実装（TDD）<br/>BE・FE"]
  I --> P["5. PR<br/>レビュー分類・CI・CodeRabbit"]
  P -->|承認・マージ| Dev["develop"]
  Dev --> Rel["6. リリース<br/>main"]
```

## 段階ごとのまとめ

| 段階 | やること | 成果物（置き場） | 次へ進む条件 | 正 |
|---|---|---|---|---|
| 1. 要求・要件 | 誰に何を伝えるために作るか（要求）と、システムと開発の進め方が満たすこと（要件）を決める | [`docs/requirements.md`](../requirements.md)・[`docs/specifications.md`](../specifications.md) | 要件が「満たしたかどうか」を確かめられる形で書かれている | 各文書 |
| 2. 設計 | 機能ごとに画面・データ・外部連携の仕様と実装計画を決める。画面が要るものは HTML モックで見え方を先に決める | 機能ごとの設計 `docs/superpowers/specs/`、画面モック `docs/design/`、実装計画 `docs/superpowers/plans/`。機能をまたぐ判断は `docs/adr/` | 設計が合意され（必要なら `review:design` の Draft PR でレビュー）、要件と食い違いが無い | [`docs/architecture.md`](../architecture.md)・[`backend/architecture.md`](../../backend/architecture.md)・[`frontend/architecture.md`](../../frontend/architecture.md)・[ADR](./adr.md) |
| 3. 詳細設計 | 業務領域・UseCase・PR の分け方を決め、API の IN・OUT・相関ルール・業務ルールの違反のコードを OpenAPI で定義する。合意したら BE・FE の生成物を作る（下記） | `openapi/<area>/<channel>/`（+ バンドル `openapi/dist/`・BE と FE の生成物） | `review:design` の Draft PR で設計レビューが OK になり、生成物を足して `develop` にマージ済み | [`docs/architecture.md`](../architecture.md) §4・[`backend/architecture.md`](../../backend/architecture.md) §3 |
| 4. 実装 | テストから書く（TDD）。BE は Domain → 残りの層の順にコミットを分ける。FE は詳細設計のマージ後に BE と並行して作れる（API は生成 MSW ハンドラで差し替える） | コードとテスト | 自動テストと CI（`Test Presence` を含む）がすべて通る | [`docs/rules/testing.md`](./testing.md)・[コメントの書き方](./code-comments.md) |
| 5. PR | レビュー分類を人が決め、Draft で作成 → CI → CodeRabbit の指摘に対応 → 確認事項を埋めて Ready → レビュー依頼 | PR | 承認 → `develop` へマージ、ブランチ削除 | [`docs/rules/git-pull-request.md`](./git-pull-request.md) |
| 6. リリース | `develop` を `main` へマージする（緊急時は `hotfix/*`） | `main` | — | [`docs/rules/git-branch.md`](./git-branch.md) |

## 1. 要求・要件

- 要求は [`docs/requirements.md`](../requirements.md)、要件は [`docs/specifications.md`](../specifications.md)。文字数などの細かい上限は、各機能の設計で決めて要件に足す
- 要件のあいまいさを実装の都合だけで決めない。決めたことは要件に書き戻す

## 2. 設計

- 機能ごとの設計は `docs/superpowers/specs/<日付>-<機能>-design.md`、実装計画は `docs/superpowers/plans/`。機能スコープの判断はその設計文書の中に書き、機能をまたぐ判断（ライブラリ・インフラ・見送ったレビュー指摘など）だけ [ADR](./adr.md) にする
- 概要設計は `docs/design/<機能>.md`（テンプレート [`docs/design/_template.md`](../design/_template.md)）。画面が要る機能は、同じディレクトリの HTML モック `docs/design/<機能>/` で見え方・文言・状態（0 件・エラー）を先に決める。入力の制約（★ 列）と相関ルール（CR-xx）・メッセージ（MSG-xx）はここで決め、詳細設計に転記する。モックのコードを本番コードへ流用しない
- 設計だけを先にレビューしたいときは、設計文書だけの Draft PR を `review:design` で出す（[PR 運用ルール](./git-pull-request.md)）

## 3. 詳細設計（OpenAPI）

API の契約を `openapi/` の YAML で定義する。これが BE・FE の生成と実装・テストの正になる（[`docs/architecture.md`](../architecture.md) §4）。

- 書き始める前に、業務領域（`internal/<area>`）と UseCase の分け方、Command（更新）か Query（参照）か、PR の分け方を決める。置き場 `openapi/<area>/<channel>/` が業務領域とチャネル（`public` / `admin`）で決まるため
- 概要設計の入出力項目一覧の ★ 列（桁数・型・必須）を `required`・`minLength`・`maxLength`・`minimum`・`maximum`・`maxItems` などに転記する。相関ルール（CR-xx）は operation の `x-correlation-rules` に「常に真になる条件」として転記し（`id` は概要設計の rule id、`message` は MSG-xx）、業務ルールの違反（対象が無い・版の不一致など）は `x-error-codes` にコード・ステータス・文言（MSG-xx か文言）・`params` を書く
- `openapi/` の差分だけを先に **Draft PR**（`review:design`）で出し、設計レビューを受ける。合意したら同じ PR で `make -C backend openapi`・`pnpm -C frontend gen:api` を実行して生成物を足し、マージする。Handler がまだ無い operation は、`cmd/api/server.go` の `var _ webapi.ServerInterface` が通るよう、`cmd/api/unimplemented.go` に `echo.ErrNotImplemented`（501）を返すスタブを置く（実装の PR でスタブを消して Handler を埋め込む。生成 IF を埋め込んで未実装を隠す形は採らない）
- CI `Test Presence` は Handler のまだ無い operation を BE のテストの検査から外すので、詳細設計の PR を先にマージできる
- 業務ルールは OpenAPI の検証だけに頼らず Domain でも保証する（[`backend/architecture.md`](../../backend/architecture.md) §3）

## 4. 実装（TDD）

- 順番は **仕様 → テスト → 実装**。テストは要件・概要設計（★ 列・相関ルール）・OpenAPI（IN・OUT・`x-correlation-rules`・`x-error-codes`）から書き、実装の今の動きから書かない（テストの書き方・置き場は [テスト規約](./testing.md)）
- BE は Domain → UseCase → Infrastructure → Presentation の順にコミットを分け、レビュアーが Domain から順に読めるようにする（業務ルールを HTTP・DB の差分に埋もれさせないため）。相関ルールは UseCase に実装し、規則ごとにテストする
- 実装中に `openapi/` の変更（と、それに伴う生成物の変更）が必要になったら、実装の PR では変えず、詳細設計に戻る（契約は BE と FE の両方が使うので、実装の PR に紛れると設計レビューを通らないため）。説明文の誤字のような小さな直しも同じにする
- FE は生成した hook を feature の hook から呼び、API の差し替えは生成 MSW ハンドラで行う。フォームの入力チェックは生成した zod スキーマを使う
- テストの有無は CI `Test Presence` が検査する（API があるのにテストが無ければ失敗、変更したソースの同じディレクトリにテストの変更が無ければ警告）
- 実装中に要件・設計と食い違いを見つけたら、コードで辻褄を合わせず、先に要件・設計を直す（戻る先は、食い違いの元になっている段階）

## 5. PR

1. AI がレビュー分類と理由を提案し、人が決定する。決定後に Draft PR を作り、分類ラベルを 1 つ付ける
2. CI をすべて通す
3. CodeRabbit の指摘を吟味して対応する（1 指摘 1 コミット、スレッドに返信）
4. PR の「確認事項」のチェックボックスをすべて埋める（ローカルで確認した内容を具体的に書く）
5. 2〜4 がすべて OK になったら Ready にしてレビューを依頼する（それより前に依頼しない）
6. 承認されたらマージし、ブランチを削除する

## 6. リリース

`develop` に集めた変更を `main` へマージして公開する。緊急修正は `main` から `hotfix/*` を切り、`main` と `develop` の両方へ反映する（[ブランチ運用ルール](./git-branch.md)）。
