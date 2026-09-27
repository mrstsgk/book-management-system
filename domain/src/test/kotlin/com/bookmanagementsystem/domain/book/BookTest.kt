package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe

class BookTest : FunSpec({
    test("有効なパラメータで書籍を作成できる") {
        val bookId = ID<Book>(1)
        val title = "吾輩は猫である"
        val price = BookPrice.of(1200L)
        val status = BookPublishStatus.PUBLISHED
        val authorIds: List<ID<Author>> = listOf(ID(1))

        val book = Book(bookId, title, price, authorIds, status)

        book.id shouldBe bookId
        book.title shouldBe title
        book.price shouldBe price
        book.status shouldBe status
        book.authorIds shouldBe authorIds
    }

    test("複数の著者を持つ書籍を作成できる") {
        val bookId = ID<Book>(2)
        val title = "共著小説"
        val price = BookPrice.of(1800L)
        val status = BookPublishStatus.UNPUBLISHED
        val authorIds: List<ID<Author>> = listOf(ID(1), ID(2), ID(3))

        val book = Book(bookId, title, price, authorIds, status)

        book.authorIds.size shouldBe 3
        book.authorIds shouldBe authorIds
    }

    test("著者が設定されていない場合、書籍を作成できない") {
        val bookId = ID<Book>(1)
        val title = "無名の書"
        val price = BookPrice.of(1000L)
        val status = BookPublishStatus.PUBLISHED
        val authorIds: List<ID<Author>> = emptyList()

        shouldThrow<IllegalArgumentException> {
            Book(bookId, title, price, authorIds, status)
        }
    }
})
