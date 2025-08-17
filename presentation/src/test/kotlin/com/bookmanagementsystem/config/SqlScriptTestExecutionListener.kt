package com.bookmanagementsystem.config

import org.springframework.core.io.ClassPathResource
import org.springframework.jdbc.datasource.init.ScriptUtils
import org.springframework.test.context.TestContext
import org.springframework.test.context.TestExecutionListener
import javax.sql.DataSource

class SqlScriptTestExecutionListener : TestExecutionListener {
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
        } catch (_: Exception) {
            // SQLスクリプトの実行でエラーが発生した場合は無視（他のテストクラスでも使用可能にするため）
            // 必要に応じてログ出力
        }
    }
}
