package com.bookmanagementsystem.config

import org.springframework.boot.jdbc.DataSourceBuilder
import org.springframework.boot.test.context.TestConfiguration
import org.springframework.context.annotation.Bean
import java.io.File
import java.io.FileInputStream
import java.util.*
import javax.sql.DataSource

@TestConfiguration
class TestConfig {

    /**
     * PostgreSQL データベースへの接続に使用する `DataSource` を生成する Bean 定義。
     *
     * このメソッドは、`.env` ファイルからデータベースの接続情報（ユーザー名とパスワード）を読み込み、
     * `DataSourceBuilder` を使って PostgreSQL 向けの `DataSource` を構築する。
     */
    @Bean
    fun dataSource(): DataSource {
        val projectRoot = System.getProperty("user.dir")
        // Check if we're in a subproject directory (presentation/), go up one level
        val rootDir = if (projectRoot.endsWith("presentation")) {
            File(projectRoot).parent
        } else {
            projectRoot
        }
        val envFile = File(rootDir, ".env")
        check(envFile.exists()) {
            "Required .env file not found at: ${envFile.absolutePath}. " +
                "Please create .env file with POSTGRES_USER and POSTGRES_PASSWORD"
        }

        val properties = Properties()
        properties.load(FileInputStream(envFile))

        val postgresUser = properties.getProperty("POSTGRES_USER")
        val postgresPassword = properties.getProperty("POSTGRES_PASSWORD")

        require(!postgresUser.isNullOrBlank() && !postgresPassword.isNullOrBlank()) {
            "POSTGRES_USER and POSTGRES_PASSWORD must be set in .env file"
        }

        return DataSourceBuilder.create()
            .url("jdbc:postgresql://localhost:5433/book_management")
            .username(postgresUser)
            .password(postgresPassword)
            .driverClassName("org.postgresql.Driver")
            .build()
    }
}
