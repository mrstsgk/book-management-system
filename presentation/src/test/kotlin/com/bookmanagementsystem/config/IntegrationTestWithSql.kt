package com.bookmanagementsystem.config

import org.springframework.test.context.TestExecutionListeners
import org.springframework.test.context.support.DependencyInjectionTestExecutionListener
import org.springframework.test.context.support.DirtiesContextTestExecutionListener
import org.springframework.test.context.transaction.TransactionalTestExecutionListener

/**
 *　### 注意事項
 *  ```
 * このアノテーションがついたテストを実行する際は以下の二点事前に実行する
 * - docker の起動
 * - Spring Boot アプリケーションを起動
 *  ```
 *
 * ### 概要
 * SQLスクリプトを伴う統合テスト用のカスタムアノテーション。
 *
 * このアノテーションをクラスに付与することで、`@IntegrationTest` に加えて、テスト実行前後に指定されたSQLスクリプトを実行できる。
 * テスト用のデータセット準備やクリーンアップを自動化する目的で使用する。
 *
 * ### 主な機能：
 * - `@IntegrationTest` を継承しており、Spring Boot の統合テスト環境を自動適用。
 *
 * ### 使用例：
 * ```
 * @IntegrationTestWithSql(sqlScript = "sql/CreateAuthorControllerTest.sql")
 * class CreateAuthorControllerTest : FunSpec({
 *     // テスト本体
 * })
 * ```
 *
 * @property sqlScript 実行するSQLスクリプトのパス（例: `"sql/SomeTest.sql"`）。空文字の場合はスクリプトなし。
 */
@Target(AnnotationTarget.CLASS)
@Retention(AnnotationRetention.RUNTIME)
@IntegrationTest
@TestExecutionListeners(
    DependencyInjectionTestExecutionListener::class,
    DirtiesContextTestExecutionListener::class,
    TransactionalTestExecutionListener::class,
    SqlScriptTestExecutionListener::class
)
annotation class IntegrationTestWithSql(
    val sqlScript: String = ""
)
