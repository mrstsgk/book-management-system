package com.bookmanagementsystem.infrastructure.config

import org.jooq.DSLContext
import org.jooq.SQLDialect
import org.jooq.impl.DSL
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.jdbc.DataSourceBuilder
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.ComponentScan
import java.io.File
import java.io.FileInputStream
import java.util.*
import javax.sql.DataSource

@SpringBootApplication
@ComponentScan(basePackages = ["com.bookmanagementsystem.infrastructure"])
class TestConfig {

    /**
     * PostgreSQL データベースへの接続に使用する `DataSource` を生成する Bean 定義。
     *
     * このメソッドは、`.env` ファイルからデータベースの接続情報（ユーザー名とパスワード）を読み込み、
     * `DataSourceBuilder` を使って PostgreSQL 向けの `DataSource` を構築する。
     */
    @Bean
    fun dataSource(): DataSource {
        val (postgresUser, postgresPassword) = getDatabaseCredentials()

        return DataSourceBuilder.create()
            .url("jdbc:postgresql://localhost:5433/book_management")
            .username(postgresUser)
            .password(postgresPassword)
            .driverClassName("org.postgresql.Driver")
            .build()
    }

    @Bean
    fun dslContext(dataSource: DataSource): DSLContext {
        return DSL.using(dataSource, SQLDialect.POSTGRES)
    }

    private fun getDatabaseCredentials(): Pair<String, String> {
        return getCredentialsForCiEnvironment() ?: getCredentialsForLocalEnvironment()
    }

    private fun getCredentialsForCiEnvironment(): Pair<String, String>? {
        val envUser = System.getenv("POSTGRES_USER")
        val envPassword = System.getenv("POSTGRES_PASSWORD")

        return if (!envUser.isNullOrBlank() && !envPassword.isNullOrBlank()) {
            Pair(envUser, envPassword)
        } else {
            null
        }
    }

    private fun getCredentialsForLocalEnvironment(): Pair<String, String> {
        val projectRoot = System.getProperty("user.dir")
        val rootDir = if (projectRoot.endsWith("infrastructure")) {
            File(projectRoot).parent
        } else {
            projectRoot
        }
        val envFile = File(rootDir, ".env")

        if (!envFile.exists()) error("Required .env file not found at: ${envFile.absolutePath}")

        val properties = Properties()
        properties.load(FileInputStream(envFile))
        val fileUser = properties.getProperty("POSTGRES_USER")
        val filePassword = properties.getProperty("POSTGRES_PASSWORD")

        require(fileUser.isNotBlank() && filePassword.isNotBlank()) {
            "POSTGRES_USER and POSTGRES_PASSWORD must be set in .env file"
        }

        return Pair(fileUser, filePassword)
    }
}
