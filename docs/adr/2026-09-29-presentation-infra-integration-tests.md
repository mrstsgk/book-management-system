# Presentation→Infrastructure の結合テストを追加する

**日付:** 2026-09-29
**状態:** 採用
**カテゴリ:** architecture
**参照:** [`docs/rules/testing.md`](../rules/testing.md)、なし（口頭決定）

## 背景・経緯

これまでのテスト方針（`docs/rules/testing.md`）は、Handler・UseCase・Infrastructure
（Repository/Query/ExternalGateway）の3層をそれぞれ独立に、Fake または契約テストで検証する形を
とっていた。UseCase のテストは明示的に「infrastructure に繋がずに走ることを目標にする」としており、
Handler のテストも Fake な UseCase を注入する形だった。

この方針には、Handler→UseCase→Repository/Query→実DB という配線全体が正しく噛み合っているかを
確認するテストが1つも存在しないという抜けがあった。各層は単体では正しくても、DI配線の間違い
（型の不一致、nilの取り違え等）は3層のどのテストにも引っかからない可能性がある。

## 決定

- 各 Handler につき、実 HTTP リクエスト・実 DI 配線・実 PostgreSQL を通した結合テストを追加する
- テストの単位は「内部の分岐網羅」ではなく「その Handler が返しうる HTTP ステータスごとに1件」とする
  （200 のハッピーパス＋実際に返しうる 4xx/5xx を1件ずつ）
- 既存の Fake ベースの Handler／UseCase テストと、Repository/Query の契約テストは維持する
  （分岐網羅・境界値の検証はそれらの責務のまま）。今回追加するテストはそれらを代替しない
- openBD・楽天など外部 API に依存する Handler は、その外部依存だけ Fake の Gateway に差し替え、
  DB は実物を使う。外部ゲートウェイ自体は `infrastructure/gateway/<name>` で既に検証済みのため、
  結合テストの実行のたびに実ネットワークへ問い合わせる必要はないと判断した

## 検討した代替案

| 案 | 概要 | 却下理由 |
|---|---|---|
| A（採用） | Handler単位で、返しうるHTTPステータスごとに実DB込みの結合テストを足す | 配線ミスを検知でき、テスト数も「ステータスの数」に収まり肥大化しない |
| B | 何もしない（現状の3層独立テストのみ） | 配線ミスを検知する手段が無いまま |
| C | E2E（実サーバ起動＋外部ネットワーク込み）を導入する | 外部APIの可用性・レート制限にテストが左右され不安定。要求定義でもデプロイ運用は対象外としており、over-engineering |
| D | 分岐網羅を結合テストでもやる | 既存の Fake ベーステスト・契約テストと重複するだけで、実行コスト（実DB接続）に見合わない |

## 判断基準（任意）

次に似た判断をするとき: 新しい Handler を追加したら、その Handler が返しうる HTTP ステータスの一覧を
先に列挙し、結合テストのケース数をそこから決める（内部の条件分岐の数からは決めない）。
