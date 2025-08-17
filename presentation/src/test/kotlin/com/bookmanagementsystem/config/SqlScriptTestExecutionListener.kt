package com.bookmanagementsystem.config

import org.slf4j.LoggerFactory
import org.springframework.core.io.ClassPathResource
import org.springframework.jdbc.datasource.init.ScriptException
import org.springframework.jdbc.datasource.init.ScriptUtils
import org.springframework.test.context.TestContext
import org.springframework.test.context.TestExecutionListener
import java.sql.SQLException
import javax.sql.DataSource

class SqlScriptTestExecutionListener : TestExecutionListener {
    companion object {
        private val logger = LoggerFactory.getLogger(SqlScriptTestExecutionListener::class.java)
    }

    /**
     * 各テストメソッドの実行前に呼び出される処理。
     *
     * `@IntegrationTestWithSql` アノテーションが付与されたテストクラスに対して、
     * 指定されたSQLスクリプトを実行する。
     *
     * この処理により以下が実行される：
     * 1. データベーススキーマやテーブル構造の準備
     * 2. 既存データのクリーンアップ
     * 3. IDシーケンスのリセット
     * 4. 一貫したテスト環境の保証
     *
     * 各テストメソッド前に実行されるため、全てのテストが独立して実行可能になる。
     */
    override fun beforeTestMethod(testContext: TestContext) {
        val testClass = testContext.testClass
        val annotation = testClass.getAnnotation(IntegrationTestWithSql::class.java)

        if (annotation != null && annotation.sqlScript.isNotEmpty()) {
            cleanupDatabaseAndResetSequences(testContext, annotation.sqlScript)
        }
    }

    private fun cleanupDatabaseAndResetSequences(testContext: TestContext, sqlScript: String) {
        try {
            val dataSource = testContext.applicationContext.getBean(DataSource::class.java)
            val resource = ClassPathResource("sql/$sqlScript")

            if (resource.exists()) {
                dataSource.connection.use { connection ->
                    ScriptUtils.executeSqlScript(connection, resource)
                }
            }
        } catch (e: ScriptException) {
            // SQLスクリプトの構文エラーや実行エラー
            logger.warn("Failed to execute SQL script '$sqlScript': ${e.message}", e)
        } catch (e: SQLException) {
            // データベース接続やSQL実行に関するエラー
            logger.warn("Database error while executing SQL script '$sqlScript': ${e.message}", e)
        } catch (e: Exception) {
            // その他の予期しないエラー
            logger.error("Unexpected error while executing SQL script '$sqlScript': ${e.message}", e)
        }
    }
}
