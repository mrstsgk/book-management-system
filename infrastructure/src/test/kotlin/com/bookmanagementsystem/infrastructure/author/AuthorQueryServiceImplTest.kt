package com.bookmanagementsystem.infrastructure.author

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.config.IntegrationTestWithSql
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import org.springframework.beans.factory.annotation.Autowired
import java.time.LocalDate

/**
 * AuthorQueryServiceImpl の統合テスト。
 *
 * 実際にデータベースに対してクエリを実行し、著者情報の取得機能を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + PostgreSQL）で実行される。
 *
 * ### 各ケースで1パス通せばOK
 *  context("正常系") - 書籍に紐づく著者が正常に取得できる：
 *  context("境界値") - 著者がいない書籍や存在しない書籍のケース：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * インフラストラクチャ層として、データベースからの正確なデータ取得を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "AuthorQueryServiceImplTest.sql")
class AuthorQueryServiceImplTest : FunSpec() {
    override fun extensions() = listOf(SpringExtension)

    @Autowired
    private lateinit var authorQueryService: AuthorQueryServiceImpl

    init {
        context("正常系") {
            test("単一の著者を持つ書籍の著者が正常に取得される") {
                val bookId = ID<Book>(1)

                val authors = authorQueryService.findByBookId(bookId)

                authors.size shouldBe 1
                authors[0].id.value shouldBe 1
                authors[0].name shouldBe "夏目漱石"
                authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))
            }

            test("複数の著者を持つ書籍の著者が正常に取得される") {
                val bookId = ID<Book>(2)

                val authors = authorQueryService.findByBookId(bookId)

                authors.size shouldBe 3

                // 著者IDでソートされていることを前提に検証
                authors[0].id.value shouldBe 1
                authors[0].name shouldBe "夏目漱石"
                authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))

                authors[1].id.value shouldBe 2
                authors[1].name shouldBe "太宰治"
                authors[1].birthDate shouldBe AuthorBirthDate(LocalDate.of(1909, 6, 19))

                authors[2].id.value shouldBe 3
                authors[2].name shouldBe "芥川龍之介"
                authors[2].birthDate shouldBe AuthorBirthDate(LocalDate.of(1892, 3, 1))
            }

            test("生年月日がnullの著者が正常に取得される") {
                val bookId = ID<Book>(3)

                val authors = authorQueryService.findByBookId(bookId)

                authors.size shouldBe 1
                authors[0].id.value shouldBe 4
                authors[0].name shouldBe "匿名作家"
                authors[0].birthDate shouldBe null
            }
        }
    }
}
