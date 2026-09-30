# 管理画面のログインはサーバー側セッション + httpOnly Cookie で守る

**日付:** 2026-09-30
**状態:** 採用
**カテゴリ:** architecture
**参照:** [設計](../superpowers/specs/2026-09-30-admin-login-design.md)、[置き換えた ADR](./2026-09-30-admin-token-from-env.md)

## 背景・経緯

書き込み系 API は、環境変数の管理者トークンを `Authorization: Bearer` で送る方式で守っていた。frontend はそのトークンを `VITE_ADMIN_TOKEN` から読んで画面のコードに埋め込むため、ビルドして公開すると誰でも読めて書き込める。「公開しない前提でだけ成り立つ」と ADR に残していたが、要求定義 §1 の「設計と実装がしっかりした人」を伝えるには、認証も脅威を特定して対策を選んだ形にしたい。

脅威モデルは「利用者は自分 1 人、ローカルで動かす、公開しない」。守るのは「書き込み系 API を自分以外に使われないこと」。

## 決定

ID + パスワードでログインし、サーバー側セッション（Postgres の `admin_session`）を httpOnly Cookie で持つ。Bearer トークン方式は残さない（認証手段を 1 つにする）。

| 脅威 | 対策 | どこで確かめるか |
|---|---|---|
| 画面のコードに埋め込まれた秘密が読まれる | 秘密（パスワードの bcrypt ハッシュ）はバックエンドの環境変数にだけ置き、ブラウザには意味を持たないセッション ID しか渡さない | `postgres/auth` の契約テスト（テーブルに秘密のカラムが無い） |
| XSS でセッションが盗まれる | Cookie を `HttpOnly` にする | `presentation/http/auth` の Handler テスト |
| CSRF | Cookie を `SameSite=Lax` にし、状態を変える要求は `Origin` が自サイトでなければ 403 | `common.RequireSameOrigin` のテスト |
| セッション ID の推測 | `crypto/rand` 32 byte を base64url | `domain/auth` のテスト |
| 盗まれたセッションが使われ続ける | アイドル 1 時間・絶対 24 時間で失効。ログアウトでサーバー側の行を消す | `domain/auth`・`usecase/auth`・結合テスト（ログアウト後の再利用が 401） |
| パスワードの総当たり | bcrypt（コスト 12）で照合を遅くし、失敗 5 回で 1 分ロック | `infrastructure/auth`・`usecase/auth/command` のテスト |
| 応答時間から ID の存在を推測される | ID が違っても bcrypt を必ず実行する。ID の比較は定数時間 | `LoginUsecase` のテスト（ID 不一致でも Verifier が呼ばれる） |
| 設定漏れで誰でも書き込める | `ADMIN_ID` / `ADMIN_PASSWORD_HASH` に既定値を置かない | `config` のテスト |

依存の追加: `golang.org/x/crypto`（bcrypt）を間接依存から直接依存に上げる。自前でハッシュを実装しないため。

## 検討した代替案

| 案 | 概要 | 却下理由 |
|---|---|---|
| A（採用） | サーバー側セッション + httpOnly Cookie | 上の表のとおり脅威ごとに対策を置け、Postgres 以外の新しいインフラが要らない |
| B | ログイン画面でトークンを入力し sessionStorage に持つ | 画面のコードに埋め込む問題は解けるが、sessionStorage は JS から読めるため XSS で即流出する。失効の手段も無い |
| C | マネージド ID プロバイダ + 外部セッションストア | 要求定義 §4（ユーザー管理なし・デプロイなし）と釣り合わない。ローカルではエミュレータ相手にしか動かず、本物に対して動いた保証が得られない |
| D | Bearer トークンのまま、トークンだけ画面で入力する | B と同じ弱点。CSRF は成立しないが XSS には無防備 |

## 判断基準（任意）

公開するときに足すもの: Cookie の `Secure`（https になるため）、frontend を別オリジンに置くなら `credentials: 'include'`、ロックの永続化（多プロセス化するなら）。ユーザーが 2 人以上になったら、この方式を捨てて C を検討する。
