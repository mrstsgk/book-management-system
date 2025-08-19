package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.config.IntegrationTestWithSql
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import java.math.BigDecimal
import java.time.LocalDate

/**
 * BookQueryServiceImpl の統合テスト。
 *
 * 実際にデータベースに対してクエリを実行し、書籍情報の取得機能を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + PostgreSQL）で実行される。
 *
 * インフラストラクチャ層として、データベースからの正確なデータ取得を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "BookDetailQueryServiceImplTest.sql")
class BookDetailQueryServiceImplTest(private val bookQueryService: BookDetailQueryService) : FunSpec() {
    override fun extensions() = listOf(SpringExtension)

    init {
        context("正常系") {
            test("単一の著者を持つ書籍が正常に取得される") {
                val bookId = ID<Book>(1)

                val book = bookQueryService.findById(bookId)

                book.id.value shouldBe 1
                book.title shouldBe "吾輩は猫である"
                book.price shouldBe BookPrice.of(BigDecimal("1500.00"))
                book.status shouldBe BookPublishStatus.PUBLISHED
                book.authors.size shouldBe 1
                book.authors[0].id.value shouldBe 1
                book.authors[0].name shouldBe "夏目漱石"
                book.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))
            }

            test("複数の著者を持つ書籍が正常に取得される") {
                val bookId = ID<Book>(2)

                val book = bookQueryService.findById(bookId)

                book.id.value shouldBe 2
                book.title shouldBe "日本文学選集"
                book.price shouldBe BookPrice.of(BigDecimal("3000.00"))
                book.status shouldBe BookPublishStatus.PUBLISHED
                book.authors.size shouldBe 3

                // 著者IDでソートされていることを前提に検証
                book.authors[0].id.value shouldBe 1
                book.authors[0].name shouldBe "夏目漱石"
                book.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))

                book.authors[1].id.value shouldBe 2
                book.authors[1].name shouldBe "太宰治"
                book.authors[1].birthDate shouldBe AuthorBirthDate(LocalDate.of(1909, 6, 19))

                book.authors[2].id.value shouldBe 3
                book.authors[2].name shouldBe "芥川龍之介"
                book.authors[2].birthDate shouldBe AuthorBirthDate(LocalDate.of(1892, 3, 1))
            }

            test("生年月日がnullの著者を持つ書籍が正常に取得される") {
                val bookId = ID<Book>(3)

                val book = bookQueryService.findById(bookId)

                book.id.value shouldBe 3
                book.title shouldBe "謎の小説"
                book.price shouldBe BookPrice.of(BigDecimal("2000.00"))
                book.status shouldBe BookPublishStatus.PUBLISHED
                book.authors.size shouldBe 1
                book.authors[0].id.value shouldBe 4
                book.authors[0].name shouldBe "匿名作家"
                book.authors[0].birthDate shouldBe null
            }

            test("存在しない書籍IDの場合は例外が発生する") {
                val bookId = ID<Book>(999)

                shouldThrow<NoSuchElementException> {
                    bookQueryService.findById(bookId)
                }
            }
        }
    }
}
