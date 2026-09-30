# 管理画面のログイン・ログアウト（サーバー側セッション）の設計

**日付:** 2026-09-30
**参照:** [要件定義 §3.1](../../specifications.md#31-ログイン)、[管理画面の設計](./2026-09-30-admin-screens-design.md)、[旧 ADR（トークンを環境変数から読む）](../../adr/2026-09-30-admin-token-from-env.md)

管理画面（`/admin/*`）にログイン・ログアウトを付け、書き込み系 API の認証を「環境変数の管理者トークンを Bearer ヘッダーで送る」方式から「サーバー側セッション + httpOnly Cookie」方式に置き換える。認証手段は 1 つにし、Bearer 方式は残さない。

## 目的と脅威モデル

要求定義 §1 のとおり、この機能で採用担当に伝えるのは「セキュリティ機能を持っている」ことではなく、**脅威を特定し、規模に合わせて対策を選び、テストで担保した**姿勢である。したがって次の 4 つを受け入れ基準にする。

1. 脅威 → 対策の対応が ADR に書いてある（対策の列挙ではなく、脅威ごとに理由が付いている）
2. 各対策にテストがある（§テスト）
3. やらなかったことと理由が書いてある（§やらないこと、ADR）
4. README から ADR と実装に辿れる

**脅威モデル**: 利用者は自分 1 人、ローカルで動かし、公開しない（要求定義 §4）。守るのは「書き込み系 API を自分以外に使われないこと」。

| 脅威 | 対策 |
|---|---|
| 画面のコードに埋め込まれた秘密が読まれる（現行方式の弱点） | 秘密（パスワードのハッシュ）はバックエンドの環境変数にだけ置く。ブラウザには意味を持たないセッション ID しか渡さない |
| XSS でセッションが盗まれる | Cookie を `HttpOnly` にし、JavaScript から読めなくする |
| CSRF（他サイトから書き込み要求を送らされる） | Cookie を `SameSite=Lax` にし、状態を変える要求では `Origin` が自サイトでなければ 403 |
| セッション ID の推測 | `crypto/rand` 32 byte を base64url にする |
| 盗まれたセッションが使われ続ける | アイドル 1 時間・絶対 24 時間で失効。ログアウトでサーバー側の行を消し、その ID は二度と通らない |
| パスワードの総当たり | bcrypt（コスト 10 以上）で照合を遅くし、失敗 5 回で 1 分間ロック |
| 応答時間から ID の存在を推測される | ID が違っても bcrypt の照合を必ず実行する。ID の比較は定数時間 |
| 設定漏れで誰でも書き込める | `ADMIN_ID` / `ADMIN_PASSWORD_HASH` に既定値を置かない。未設定ならログインできない |

## API（backend）

すべて `/api/auth` 配下。`POST` の 2 本には Origin 検証（`RequireSameOrigin`）を掛ける。

| Method | Path | 役割 | 応答 |
|---|---|---|---|
| `POST` | `/api/auth/login` | `{ id, password }` を照合し、セッションを発行して `Set-Cookie` | 204 / 400（空欄）/ 401（不一致）/ 429（ロック中） |
| `POST` | `/api/auth/logout` | セッション行を削除し、Cookie を消す（`Max-Age=0`） | 204（未ログインでも 204） |
| `GET` | `/api/auth/session` | ログイン中か（画面のガード用）。有効ならアイドル期限を延長する | 204 / 401 |

- 401 の本文はどちらが違うかを出さない（ログインは `unauthorized: IDかパスワードが違います`、未ログインは `…ログインしてください`、期限切れは `…ログインの期限が切れました`）
- swag の `@securityDefinitions.apikey AdminToken` は削除する（Swagger 2.0 は Cookie 認証を表せない）。認証が要る API は `@Security` の代わりに `@Description` に「ログイン済みの Cookie が要る」と書く。Orval は securityDefinitions を生成に使わないので、生成物への影響は無い

### Cookie

`admin_session=<ID>; HttpOnly; SameSite=Lax; Path=/api; Max-Age=86400`。`Secure` は付けない（ローカルの http でしか動かさない。公開するときに足す。§やらないこと）。

### セッションの保存

Postgres に `admin_session` テーブルを追加する（マイグレーション `000008`、`docs/db/backend-schema.{json,md}` を同じ変更で更新）。

| カラム | 型 | 意味 |
|---|---|---|
| `id` | `VARCHAR(64)` PK | セッション ID（base64url、43 文字） |
| `expires_at` | `TIMESTAMPTZ` NOT NULL | アイドル期限。管理者向け API（`GET /session` と `GET /catalog/:isbn` を含む）を通るたびに「今 + 1 時間」に延ばす（絶対期限を超えない） |
| `absolute_expires_at` | `TIMESTAMPTZ` NOT NULL | 絶対期限。発行時刻 + 24 時間で固定 |
| `created_at` | `TIMESTAMPTZ` NOT NULL DEFAULT `NOW()` | 発行日時 |

期限切れの行は消さずに残す（判定は `expires_at` / `absolute_expires_at` の比較。1 人利用で行数は増えない）。パスワードもハッシュもテーブルには入れない。

### 照合と総当たり

- `ADMIN_ID`、`ADMIN_PASSWORD_HASH`（bcrypt）を環境変数から読む。既定値は置かない。どちらかが空なら起動時にエラーにはせず、ログインが常に 401 になる（`RequireAdminToken` が空トークンを常に拒否していたのと同じ考え）
- ハッシュの生成は `go run ./cmd/hashpw`（標準入力からパスワードを読み、ハッシュを標準出力に出す。引数で受けると shell の履歴に残るため）。手順は `backend/README.md` に書く
- ID は `subtle.ConstantTimeCompare`、パスワードは `bcrypt.CompareHashAndPassword`。ID が不一致でも bcrypt を実行する
- 総当たり: `LoginUsecase` がプロセス内メモリ（自身のフィールド）で失敗回数を数え、5 回目の失敗から 1 分間は照合せずに `ErrTooManyAttempts`（429）。成功で 0 に戻す。専用の型は作らない。再起動で消える・多プロセス非対応は `ponytail:` コメントと ADR に明記する

### 層の置き場（`backend/architecture.md` §2 に従う）

| 層 | 追加・変更 |
|---|---|
| `domain/common` | sentinel `ErrUnauthorized`、`ErrTooManyAttempts` を追加 |
| `domain/auth` | `Credentials` VO（ID 空・パスワード空を拒否。トリムしない）、`Session` Entity（`IsValid(now)` はアイドル期限との 1 比較。`Extend(now)` が絶対期限で頭打ちにするので、アイドル期限は常に絶対期限以下）、`SessionRepository` IF（Save / Find / Delete / UpdateExpiry）、`PasswordVerifier` IF（ExternalGateway 相当。bcrypt を Domain から隠す） |
| `usecase/auth/command` | `LoginUsecase`（`LoginCommand{ID, Password}` → セッション IDを返す。失敗回数とロック期限は自身のフィールド）、`LogoutUsecase`（ID を scalar で受ける） |
| `usecase/auth/query` | `CheckSessionUsecase`（`Execute(ctx, id)` で有効判定と延長。無効なら `ErrUnauthorized`） |
| `infrastructure/postgres/auth` | `SessionRepository` 実装（GORM、`model.go` は永続化専用） |
| `infrastructure/auth` | `PasswordVerifier` の bcrypt 実装 |
| `presentation/http/common` | `RequireAdminToken` を削除し `RequireAdminSession`（Cookie → `CheckSessionUsecase`）と `RequireSameOrigin`（`Origin` が自サイトでなければ 403）を追加。`HTTPErrorHandler` に `ErrUnauthorized → 401`、`ErrTooManyAttempts → 429` を追加 |
| `presentation/http/auth` | `Handler`（login / logout / session）と `LoginRequest` DTO |
| `config` | `ADMIN_TOKEN` を削除し `ADMIN_ID` / `ADMIN_PASSWORD_HASH` を追加 |
| `cmd/api/main.go` | DI に auth を追加。`adminOnly` は `RequireAdminSession` に差し替え、書き込み系ルートには `RequireSameOrigin` も重ねる |
| `cmd/hashpw` | ハッシュ生成の小さなコマンド |

`RequireSameOrigin` は `Origin` ヘッダーが自サイト（`Host` と同じスキーム・ホスト・ポート）と一致すれば通し、違うか無ければ 403（ブラウザは POST に必ず `Origin` を付けるので `Referer` は見ない）。`GET` には掛けない。Vite の proxy は `changeOrigin: false` にして `Host` を書き換えない。

## 画面（frontend）

| 画面 | path | 使う API |
|---|---|---|
| ログイン | `/admin/login` | `POST /api/auth/login` |
| 管理画面の親（ガード） | `/admin/*` | `GET /api/auth/session` |

- ログイン画面は ID 欄・パスワード欄（`type="password"`、`autocomplete="current-password"`）・「ログイン」ボタン。空欄があれば送信しない。401 は「ID かパスワードが違います」、429 は「しばらく待ってからやり直してください」、それ以外は既存の共通エラー表示。どちらが違うかは出さない
- 成功したら、ログイン画面に来る前にいた管理画面（無ければ `/admin`）へ戻る。行き先は `useLocation().state.from` で渡す
- `AdminLayout` のヘッダーに「ログアウト」ボタンを置く。押すと `POST /api/auth/logout` を呼び、成功したら `/admin/login` へ。失敗したら（サーバー側のセッションと Cookie が残っているため）画面には留まり、ヘッダーの下にエラーを出して再試行できるようにする。既存の「トークンが設定されていません」バナーは撤去する
- ルートガード: `/admin/*` の親で `GET /api/auth/session` を呼ぶ。読み込み中は既存の読み込み表示、204 以外なら `/admin/login` へ（`state.from` に現在地）。エラー用の状態は持たない（サーバーが落ちていればログインも失敗し、その文言で伝わる）。`/admin/login` はガードの外に置く
- `lib/admin-auth.ts`（`adminToken()` / `adminRequest()`）と `VITE_ADMIN_TOKEN` は削除する。10 箇所の hook は `request` を渡さない形に戻す。Cookie は同一オリジン（Vite のプロキシ経由）なので自動で送られる。`mutator.ts` は変えない（`credentials: 'include'` は `VITE_API_BASE_URL` を別オリジンにするときだけ要る。§やらないこと）
- 書き込み中に 401 が返ったら（セッション切れ）、その hook の `onError` で `/admin/login` へ送る。3 hook（`useBookRegistration` / `useBookEditor` / `useAdminTags`）に各 1 行。共通の `onError` を差し込む仕組みは作らない
- 生成物: swag を再排出し `pnpm gen:api` で `auth` の生成フック・型・MSW ハンドラを作る。ログイン・ログアウト・セッション確認はすべて生成フックを使う

## テスト

`docs/rules/testing.md` と `frontend/architecture.md` §8 に従う。狙いは §目的の受け入れ基準 2「各対策にテストがある」。

### backend

| 層 | 対象 | 期待する振る舞い |
|---|---|---|
| domain | `Credentials` | ID 空・パスワード空はエラー。前後の空白は落とさない |
| domain | `Session` | 有効 / アイドル期限切れ（境界: ちょうど・+1 秒）。`Extend` しても絶対期限を超えない |
| usecase | `LoginUsecase`（Fake） | ID 不一致 → `ErrUnauthorized`、Repository は呼ばれない、`PasswordVerifier` は**呼ばれる**。パスワード不一致 → 同様。成功 → Repository に保存され ID が返る。5 回失敗 → 6 回目は照合せず `ErrTooManyAttempts`。ロック中は正しいパスワードでも 429 でセッションを作らない。成功で回数が戻る。Repository のエラーはそのまま伝播 |
| usecase | `LogoutUsecase` | Delete が呼ばれる。存在しない ID でもエラーにしない |
| usecase | `CheckSessionUsecase` | 期限切れ → `ErrUnauthorized`、延長は呼ばれない。有効 → 延長される |
| infrastructure | `SessionRepository` 契約テスト（実 DB） | Save / Find / Delete / UpdateExpiry。テーブルにパスワード・ハッシュのカラムが無い |
| infrastructure | bcrypt 実装 | 正しい / 誤ったパスワード。生成したハッシュのコストが 10 以上 |
| presentation | `RequireAdminSession` | Cookie 無し / 不正 ID / 期限切れ → 401。有効 → 通過 |
| presentation | `RequireSameOrigin` | Origin 一致 → 通過。不一致・`null`・無し → 403。`GET` には掛からない |
| presentation | auth `Handler`（`httptest` + `common.NewEcho()`） | login 204 の `Set-Cookie` に `HttpOnly` / `SameSite=Lax` / `Path=/api` があり、値は 43 文字の base64url。400 / 401 / 429 の変換。logout の `Set-Cookie` に `Max-Age=0` |
| 結合（実 DB、`*_integration_test.go`） | login → 書き込み API 200 → logout → 同じ Cookie で書き込み → 401 | ログアウト後の再利用が実物同士の配線で拒否される |

### frontend

| 対象 | 期待する振る舞い |
|---|---|
| `LoginPage` | 空欄で送信されない。401 / 429 の文言。成功で `state.from` へ、無ければ `/admin` へ |
| ルートガード | `session` が 401（または 500）→ `/admin/login` に遷移し `state.from` に元の path。204 → 子画面が出る。読み込み中の表示 |
| `AdminLayout` | ログアウト押下で `POST /api/auth/logout` が呼ばれ `/admin/login` へ。失敗（500）なら管理画面に留まり、エラーが出て、もう一度押せる |
| 既存 hook テスト | Bearer ヘッダーの検証を削除。書き込みが 401 → `/admin/login` |

### E2E

`helpers.ts` に `login(page)` を足し、既存 2 導線の `beforeEach` で UI からログインする。ID とパスワードは `process.env.E2E_ADMIN_ID` / `E2E_ADMIN_PASSWORD` から読む（`README` に手順）。ログイン単独の導線は足さない（画面テストで足りる）。

## docs の更新

- `docs/specifications.md` §3.1: 「後回し」を外し、ID + パスワードでログインすること・セッションの期限を書く
- `backend/architecture.md`: 認証の行と API 表を Cookie セッションに改め、`/api/auth` の 3 本を足す
- `backend/README.md` / `frontend/README.md`: `ADMIN_TOKEN` / `VITE_ADMIN_TOKEN` を `ADMIN_ID` / `ADMIN_PASSWORD_HASH`（生成手順）/ E2E 用の環境変数に改める
- `docs/superpowers/specs/2026-09-30-admin-screens-design.md`: 「ログインは後回し」節の先頭に本書で置き換えた旨を 1 行足す
- `docs/adr/2026-09-30-admin-token-from-env.md`: 状態を「不採用（本 ADR で置き換え）」にし、新 ADR へリンク
- 新 ADR `docs/adr/2026-09-30-admin-login-with-server-side-session.md`（カテゴリ architecture）: §脅威モデルの表、Bearer から Cookie に変えた理由、採らなかった案（外部の ID プロバイダ、外部のセッションストア、トークンを画面で入力する方式）とその理由、公開する場合の昇格パス（Cookie の `Secure`、`credentials: 'include'`、ロックの永続化）
- `README.md`（リポジトリ直下）: 設計の案内に ADR と `presentation/http/common` への導線を足す

## やらないこと

理由は ADR にも書く。コード側は `ponytail:` コメントで天井と昇格先を示す。

- 多重セッションの一覧・強制ログアウト（1 人利用で必要になる場面が無い）
- パスワード変更の画面（環境変数を書き換えれば足りる）
- ロックの永続化・多プロセス対応（再起動で消える。ローカル 1 プロセスの前提）
- `mutator.ts` への `credentials: 'include'`（同一オリジンのため不要。別オリジンに置くときに足す）
- Cookie の `Secure` とそれを切り替える設定（http でしか動かさない。https で公開するときに足す）
- ルートガードのエラー表示と再試行（確認に失敗したらログイン画面へ送るだけ）
- `Origin` が無いときの `Referer` フォールバック（ブラウザは POST に必ず `Origin` を付ける）
- 期限切れセッション行の定期削除（1 人利用で行数は増えない）
- 外部の ID プロバイダ・外部のセッションストアの導入（要求定義 §4 の「ユーザー管理なし・デプロイなし」と釣り合わない）
