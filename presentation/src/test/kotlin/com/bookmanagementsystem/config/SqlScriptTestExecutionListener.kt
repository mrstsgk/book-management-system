package com.bookmanagementsystem.config

import org.springframework.core.io.ClassPathResource
import org.springframework.jdbc.datasource.init.ScriptUtils
import org.springframework.test.context.TestContext
import org.springframework.test.context.TestExecutionListener
import javax.sql.DataSource

class SqlScriptTestExecutionListener : TestExecutionListener {
    /**
     * テストクラスの実行前に呼び出される処理。
     *
     * `@IntegrationTestWithSql` アノテーションが付与されたテストクラスに対して、
     * 指定された SQL スクリプト（`sqlScript`）を実行する。
     *
     * スクリプトは `resources/sql/` ディレクトリ内に存在する前提で、クラスパスから読み込まれる。
     * 実行はトランザクション外で行われ、テスト用のデータセット投入などに利用される。
     *
     */
    override fun beforeTestClass(testContext: TestContext) {
        val testClass = testContext.testClass
        val annotation = testClass.getAnnotation(IntegrationTestWithSql::class.java)

        if (annotation != null && annotation.sqlScript.isNotEmpty()) {
            val dataSource = testContext.applicationContext.getBean(DataSource::class.java)
            val resource = ClassPathResource("sql/${annotation.sqlScript}")

            if (resource.exists()) {
                dataSource.connection.use { connection ->
                    ScriptUtils.executeSqlScript(connection, resource)
                }
            }
        }
    }
}