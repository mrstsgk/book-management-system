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

- **言語**: Kotlin
- **Java バージョン**: Java 21（OpenJDK/Amazon Corretto）
- **フレームワーク**: Spring Boot
- **テスト**: Kotest, MockK
- **データベース**: PostgreSQL
- **ORM**: jOOQ
- **マイグレーション**: Flyway
- **ビルドツール**: Gradle - Groovy
- **静的解析**: Detekt
- **開発環境**: Docker Compose（PostgreSQL）
- **利用ツール**: Spring initializr（ https://start.spring.io/ ）

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

### 4. データベース設定の確認

`src/main/resources/application.yml`でデータベース接続設定を確認：
- ユーザー名/パスワードは`.env`ファイルの値を直接指定

### 5. jOOQコード生成（初回のみ）

データベーススキーマからjOOQのコードを生成：

```bash
./gradlew :infrastructure:generateJooq
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

## 開発コマンド

```bash
# アプリケーション起動
./gradlew bootRun

# テスト実行
./gradlew test

# ビルド
./gradlew build

# jOOQコード生成
./gradlew :infrastructure:generateJooq
```

## データベース接続

- **URL**: `jdbc:postgresql://localhost:5433/book_management`
- **ユーザー名/パスワード**: `.env`ファイルで設定