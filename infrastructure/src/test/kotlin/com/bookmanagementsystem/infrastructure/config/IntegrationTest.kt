package com.bookmanagementsystem.infrastructure.config

import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.TestPropertySource

/**
 * インフラストラクチャ層の統合テスト用のカスタムアノテーション。
 *
 * このアノテーションをクラスに付与することで、Spring Boot の統合テストとして実行される。
 * 以下の設定が事前に適用される：
 *
 * - `@SpringBootTest(classes = [TestConfig::class])` を指定し、統合テスト専用の設定クラスを読み込む。
 */
@Target(AnnotationTarget.CLASS)
@Retention(AnnotationRetention.RUNTIME)
@SpringBootTest(classes = [TestConfig::class])
@TestPropertySource(
    properties = [
        "spring.jpa.hibernate.ddl-auto=none",
        "spring.sql.init.mode=never"
    ]
)
annotation class IntegrationTest