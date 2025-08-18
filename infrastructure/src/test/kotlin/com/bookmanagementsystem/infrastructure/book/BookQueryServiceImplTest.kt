package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.config.IntegrationTestWithSql
import com.bookmanagementsystem.usecase.book.read.BookQueryService
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import java.math.BigDecimal
import java.time.LocalDate

/**
 * BookQueryServiceImpl の統合テスト。
 *
 * 実際にデータベースに対してクエリを実行し、著者IDに紐づく書籍一覧の取得機能を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + PostgreSQL）で実行される。
 *
 * インフラストラクチャ層として、データベースからの正確なデータ取得を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "BookQueryServiceImplTest.sql")
class BookQueryServiceImplTest(private val bookQueryService: BookQueryService) : FunSpec() {
    override fun extensions() = listOf(SpringExtension)

    init {
        context("正常系") {
            test("著者IDに紐づく書籍が正常に取得される") {
                val authorId = ID<Author>(1)

                val books = bookQueryService.findByAuthorId(authorId)

                books.size shouldBe 2

                // 書籍IDでソート済みであることを前提に検証
                val book1 = books[0]
                book1.id.value shouldBe 1
                book1.title shouldBe "吾輩は猫である"
                book1.price shouldBe BookPrice.of(BigDecimal("1500.00"))
                book1.status shouldBe BookPublishStatus.PUBLISHED
                book1.authors.size shouldBe 1
                book1.authors[0].id.value shouldBe 1
                book1.authors[0].name shouldBe "夏目漱石"
                book1.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))

                val book2 = books[1]
                book2.id.value shouldBe 2
                book2.title shouldBe "日本文学選集"
                book2.price shouldBe BookPrice.of(BigDecimal("3000.00"))
                book2.status shouldBe BookPublishStatus.PUBLISHED
                book2.authors.size shouldBe 3
            }

            test("著者IDに紐づく複数の書籍が正常に取得される") {
                val authorId = ID<Author>(2)

                val books = bookQueryService.findByAuthorId(authorId)

                books.size shouldBe 2

                // 書籍IDでソート済みであることを前提に検証
                val book1 = books[0]
                book1.id.value shouldBe 2
                book1.title shouldBe "日本文学選集"
                book1.price shouldBe BookPrice.of(BigDecimal("3000.00"))
                book1.status shouldBe BookPublishStatus.PUBLISHED
                book1.authors.size shouldBe 3

                val book2 = books[1]
                book2.id.value shouldBe 4
                book2.title shouldBe "文学論"
                book2.price shouldBe BookPrice.of(BigDecimal("2500.00"))
                book2.status shouldBe BookPublishStatus.PUBLISHED
                book2.authors.size shouldBe 1
                book2.authors[0].id.value shouldBe 2
                book2.authors[0].name shouldBe "太宰治"
                book2.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1909, 6, 19))
            }

            test("共著書籍の場合、すべての著者情報が正常に取得される") {
                val authorId = ID<Author>(1)

                val books = bookQueryService.findByAuthorId(authorId)
                val collaborativeBook = books.find { it.id.value == 2 }

                collaborativeBook!!.authors.size shouldBe 3

                // 著者IDでソートされていることを前提に検証
                collaborativeBook.authors[0].id.value shouldBe 1
                collaborativeBook.authors[0].name shouldBe "夏目漱石"
                collaborativeBook.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1867, 2, 9))

                collaborativeBook.authors[1].id.value shouldBe 2
                collaborativeBook.authors[1].name shouldBe "太宰治"
                collaborativeBook.authors[1].birthDate shouldBe AuthorBirthDate(LocalDate.of(1909, 6, 19))

                collaborativeBook.authors[2].id.value shouldBe 3
                collaborativeBook.authors[2].name shouldBe "芥川龍之介"
                collaborativeBook.authors[2].birthDate shouldBe AuthorBirthDate(LocalDate.of(1892, 3, 1))
            }

            test("著者IDに紐づく単一の書籍が正常に取得される") {
                val authorId = ID<Author>(3)

                val books = bookQueryService.findByAuthorId(authorId)

                books.size shouldBe 2

                // 書籍IDでソート済みであることを前提に検証（共著書籍と単独著作）
                val collaborativeBook = books[0]
                collaborativeBook.id.value shouldBe 2
                collaborativeBook.title shouldBe "日本文学選集"
                collaborativeBook.authors.size shouldBe 3

                val singleAuthorBook = books[1]
                singleAuthorBook.id.value shouldBe 5
                singleAuthorBook.title shouldBe "こころ"
                singleAuthorBook.price shouldBe BookPrice.of(BigDecimal("1800.00"))
                singleAuthorBook.status shouldBe BookPublishStatus.PUBLISHED
                singleAuthorBook.authors.size shouldBe 1
                singleAuthorBook.authors[0].id.value shouldBe 3
                singleAuthorBook.authors[0].name shouldBe "芥川龍之介"
                singleAuthorBook.authors[0].birthDate shouldBe AuthorBirthDate(LocalDate.of(1892, 3, 1))
            }

            test("生年月日がnullの著者を持つ書籍が正常に取得される") {
                val authorId = ID<Author>(4)

                val books = bookQueryService.findByAuthorId(authorId)

                books.size shouldBe 1
                val book = books[0]
                book.id.value shouldBe 3
                book.title shouldBe "謎の小説"
                book.price shouldBe BookPrice.of(BigDecimal("2000.00"))
                book.status shouldBe BookPublishStatus.PUBLISHED
                book.authors.size shouldBe 1
                book.authors[0].id.value shouldBe 4
                book.authors[0].name shouldBe "匿名作家"
                book.authors[0].birthDate shouldBe null
            }

            test("存在しない著者IDの場合は空のリストが返される") {
                val authorId = ID<Author>(999)

                val books = bookQueryService.findByAuthorId(authorId)

                books shouldBe emptyList()
            }

            test("書籍を持たない著者IDの場合は空のリストが返される") {
                val authorId = ID<Author>(5)

                val books = bookQueryService.findByAuthorId(authorId)

                books shouldBe emptyList()
            }
        }
    }
}
