# アーキテクチャ方針

自分が読んだ本を紹介するポートフォリオ。バックエンド API（Go）と Web 画面（React）を同一リポジトリで扱う。読んだ本は ISBN で登録し、書誌と書影は外部カタログ（openBD）から取得して、自分の感想と評価を添える。閲覧は誰でも、登録・更新・削除は自分だけができる。

## 文書の言語

| 種類 | 言語 |
|---|---|
| 人が読む方針（本書・`docs/**`・パッケージ README など） | **日本語** |
| AI 入口（`CLAUDE.md`・`.cursor/rules/*.mdc`） | **英語**（長い方針はここに複製せず、日本語の正を参照） |

| 文書 | 管轄（重複しない） |
|---|---|
| [`README.md`](../README.md) | 人が読む資料の入口（読む順番と各資料の役割）・リポジトリ全体のディレクトリ構成。規則は書かない |
| [`requirements.md`](./requirements.md) | 要求定義（誰に何を伝えるために作るか） |
| [`specifications.md`](./specifications.md) | 要件定義（システムと開発の進め方が満たすこと） |
| **本書** | リポジトリ全体・トップレベル構成・契約の**境界** |
| [`backend/architecture.md`](../backend/architecture.md) | バックエンドのスタック・層・HTTP/OpenAPI・ディレクトリ・永続化 |
| [`frontend/architecture.md`](../frontend/architecture.md) | フロントのスタック・設計パターン・アプリ内ディレクトリ・共有ルール |
| [`rules/`](./rules/) | 開発の規則（開発の流れ・Git・テスト・コメント・ADR・DB 資料） |

フロントの実装詳細は本書に書かない。バックエンド実装詳細は [`backend/architecture.md`](../backend/architecture.md) に書き、本書は境界のみ。

## 1. 決定（要約）

| 項目 | 決定 |
|---|---|
| バックエンド | **Go / Echo モノリス**（1 `go.mod`・API バイナリ1つ）。Presentation → UseCase → Domain ← Infrastructure。各層はドメイン別ディレクトリ（+ common）。CQRS で Repository／Query 分離（単一 DB）。正は [`backend/architecture.md`](../backend/architecture.md) |
| フロントエンド | **React + Vite + TypeScript** の管理画面 `frontend/web`。**mise + pnpm**。正は [`frontend/architecture.md`](../frontend/architecture.md) |
| API 契約（境界） | **BE は Go 先行**（DTO + validator + swag）。**排出 OpenAPI → FE の TypeScript 生成は必須**。生成物は手編集しない。CI でドリフト検知（gen → diff）。詳細は backend 文書 |
| モジュラモノリス / マイクロサービス | **採用しない**（境界が必要になったら後からパッケージを育てる） |
| 旧実装 | Kotlin / Spring Boot 版、および旧仕様（書籍・著者の管理 API）の Go 版は削除済み（[ADR](./adr/2026-09-28-rebuild-as-reading-portfolio.md)） |

## 2. バックエンド（境界のみ）

正は [`backend/architecture.md`](../backend/architecture.md)。

- API の path は `/api/...`（例: `/api/books`）。閲覧は認証なし、書き込みは管理者トークン
- ローカル開発: **ホスト Go（mise で版固定）+ Docker は DB**。API の日常起動はコンテナにしない（詳細は backend 文書）

## 3. フロントエンド（境界のみ）

正は [`frontend/architecture.md`](../frontend/architecture.md)。

- アプリ: `frontend/web`（読んだ本の紹介画面と、自分用の登録・編集画面）
- アプリ横断コード: `frontend/packages/`（中身の規約は frontend 文書）

## 4. OpenAPI / 型契約（境界）

| 側 | 正 | 流れ |
|---|---|---|
| Backend | Go の DTO / Handler | swag で OpenAPI を**排出** |
| Frontend | 排出された OpenAPI | **TypeScript 生成必須**（Orval） |

- リポジトリルートに手書き `openapi/` master を置き oapi-codegen で Server を生成する方式は**採らない**（詳細は backend 文書）
- **生成物（swag 排出・TS）は手編集しない**
- FE で API 型を手書きして正としない

### ドリフト検知（CI・必須）

CI で OpenAPI 排出および TS 生成を再実行 → `git diff --exit-code`（差分があれば **fail**）。

### 名前を揃えるもの・揃えないもの

目的が同じものだけ名前を揃える。FE の画面・独自の hook は目的が違うので、BE の UseCase に揃えない。

| もの | 目的 | 名前 |
|---|---|---|
| FE の画面・独自の hook | 画面で何を見せ、何をさせるか | 画面の目的で付ける（`BooksListView`・`useBooksList`。[`frontend/architecture.md`](../frontend/architecture.md) §9） |
| FE の生成 hook・関数 | その API を呼ぶ | Orval の命名（HTTP メソッド + パス。`useGetApiBooks`）のまま。swag の `operationId` は今は付けていないため、付けるなら BE の UseCase と揃える |
| BE の Handler のメソッド・UseCase | 業務の目的 | `<動詞><業務対象>`（`SearchBooksUsecase`。[`backend/architecture.md`](../backend/architecture.md) §2） |

## 5. リポジトリ構成

ディレクトリ構成の正はリポジトリ直下の [`README.md`](../README.md)「ディレクトリ構成」。要点だけ:

```
frontend/   # web/（アプリ）と packages/（共有）。詳細は frontend/architecture.md
backend/    # 詳細は backend/architecture.md（migrations SQL の正もここ）
docs/       # 人が読む方針・規則・ADR・設計
.github/    # CI・PR テンプレート
```
