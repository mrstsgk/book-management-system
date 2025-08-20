# Book Management System

書籍管理システムのSpring Boot + Kotlinアプリケーション

## アーキテクチャ

このプロジェクトはマルチモジュール構成で、クリーンアーキテクチャを採用しています：

```
book-management-system/
├── domain/                         # ドメイン層（Entity / Repository Interface）
├── infrastructure/                 # インフラ層（外部接続 / Repository Impl / jOOQ、Flyway関連のファイル）
├── presentation/                   # プレゼンテーション層（Controller）
└── usecase/                        # アプリケーション層（Service）
```

## 技術スタック

- **言語**: Kotlin 2.2.10
- **Java バージョン**: Java 21（OpenJDK/Amazon Corretto）
- **フレームワーク**: Spring Boot 3.4.2
- **テスト**: Kotest 5.9.1, MockK 1.13.13
- **データベース**: PostgreSQL 17.5
- **ORM**: jOOQ 3.19.15
- **マイグレーション**: Flyway Core
- **ビルドツール**: Gradle 8.14.3 (Groovy)
- **静的解析**: Detekt 1.23.8
- **開発環境**: Docker Compose（PostgreSQL）
- **API仕様**: OpenAPI 3.0.3
- **バリデーション**: Jakarta Bean Validation

## アプリケーション起動前の準備

### 前提条件
- Java 21以上
- Docker & Docker Compose
- Git

### 1. リポジトリのクローン
```bash
git clone https://github.com/mrstsgk/book-management-system.git
cd book-management-system
```

### 2. 環境変数の設定

`.env.sample`を参考に`.env`ファイルを作成：

```bash
cp .env.sample .env
```

`.env`ファイルを編集してPostgreSQLの認証情報を設定：
```bash
# PostgreSQL Docker Configuration
POSTGRES_USER=sample_user
POSTGRES_PASSWORD=sample_password
```

### 3. データベース（PostgreSQL）の起動

Docker Composeを使用してPostgreSQLコンテナを起動：

```bash
docker compose up -d
```

コンテナが正常に起動したことを確認：
```bash
docker compose ps
```

### 5. jOOQコード生成（初回のみ）

データベーススキーマからjOOQのコードを生成：

```bash
./gradlew :infrastructure:jooqCodege
```

### 6. アプリケーションのビルド

依存関係のダウンロードとビルド：

```bash
./gradlew build
```

### 7. アプリケーションの起動

```bash
./gradlew bootRun
```

アプリケーションが起動すると、以下のURLでアクセス可能：
- アプリケーション: http://localhost:8080

## API仕様

### API一覧

#### 著者API

| メソッド | エンドポイント | 説明 | 備考 |
|---------|---------------|------|------|
| POST | `/api/authors` | 著者を作成する | 生年月日は現在より過去日付 |
| PUT | `/api/authors/{id}` | 著者を更新する | 楽観的ロック対応 |
| GET | `/api/authors/{id}/books` | 著者に紐づく書籍一覧を取得する | 書籍が見つからない場合は空配列を返す |

#### 書籍API

| メソッド | エンドポイント | 説明 | 備考 |
|---------|---------------|------|------|
| POST | `/api/books` | 書籍を作成する | 価格は0以上、著者は1人以上 |
| PUT | `/api/books/{id}` | 書籍を更新する | 出版済みから未出版に変更不可、楽観的ロック対応 |

### APIドキュメント

詳細なAPI仕様は以下のファイルで確認できます：
- OpenAPI仕様書: `presentation/src/main/resources/openapi.yml`

## データベース

### 接続情報
- **URL**: `jdbc:postgresql://localhost:5433/book_management`
- **ユーザー名/パスワード**: `.env`ファイルで設定

### テーブル構成
- **book**: 書籍情報（id, title, price, publish_status, version）
- **author**: 著者情報（id, name, birth_date, version）
- **author_book**: 著者と書籍の関連テーブル（author_id, book_id, version）

### マイグレーション
Flywayによる自動マイグレーション対応。アプリケーション起動時に自動実行されます。