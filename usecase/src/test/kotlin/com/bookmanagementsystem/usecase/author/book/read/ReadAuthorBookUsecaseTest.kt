package com.bookmanagementsystem.usecase.author.book.read

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.book.BookSummaryDto
import com.bookmanagementsystem.usecase.book.read.BookQueryService
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.math.BigDecimal
import java.time.LocalDate

class ReadAuthorBookUsecaseTest : FunSpec({
    test("著者IDに紐づく書籍一覧が正常に取得されること") {
        val authorRepository = mockk<AuthorRepository>()
        val bookQueryService = mockk<BookQueryService>()
        val usecase = ReadAuthorBookUsecase(authorRepository, bookQueryService)
        val authorId = ID<Author>(1)
        val author = Author(
            id = authorId,
            name = "夏目漱石",
            birthDate = AuthorBirthDate(LocalDate.of(1867, 2, 9))
        )
        val expectedBooks = listOf(
            BookSummaryDto(
                id = ID(1),
                title = "吾輩は猫である",
                price = BookPrice.of(BigDecimal("1500.00")),
                status = BookPublishStatus.PUBLISHED
            ),
            BookSummaryDto(
                id = ID(2),
                title = "こころ",
                price = BookPrice.of(BigDecimal("1800.00")),
                status = BookPublishStatus.PUBLISHED
            )
        )

        every { authorRepository.findById(authorId) } returns author
        every { bookQueryService.findByAuthorId(authorId) } returns expectedBooks

        val result = usecase.execute(authorId)

        result shouldBe expectedBooks
        verify { authorRepository.findById(authorId) }
        verify { bookQueryService.findByAuthorId(authorId) }
    }

    test("著者に紐づく書籍が存在しない場合、空のリストが返されること") {
        val authorRepository = mockk<AuthorRepository>()
        val bookQueryService = mockk<BookQueryService>()
        val usecase = ReadAuthorBookUsecase(authorRepository, bookQueryService)
        val authorId = ID<Author>(999)
        val author = Author(
            id = authorId,
            name = "書籍なし著者",
            birthDate = AuthorBirthDate(LocalDate.of(1900, 1, 1))
        )

        every { authorRepository.findById(authorId) } returns author
        every { bookQueryService.findByAuthorId(authorId) } returns emptyList()

        val result = usecase.execute(authorId)

        result shouldBe emptyList()
        verify { authorRepository.findById(authorId) }
        verify { bookQueryService.findByAuthorId(authorId) }
    }

    test("存在しない著者IDの場合、NoSuchElementExceptionが発生すること") {
        val authorRepository = mockk<AuthorRepository>()
        val bookQueryService = mockk<BookQueryService>()
        val usecase = ReadAuthorBookUsecase(authorRepository, bookQueryService)
        val authorId = ID<Author>(999)

        every { authorRepository.findById(authorId) } returns null

        val exception = shouldThrow<NoSuchElementException> {
            usecase.execute(authorId)
        }

        exception.message shouldBe "著者が見つかりません。ID: 999"
        verify { authorRepository.findById(authorId) }
        verify(exactly = 0) { bookQueryService.findByAuthorId(any()) }
    }

    test("複数の著者による共著書籍が含まれる場合も正常に取得されること") {
        val authorRepository = mockk<AuthorRepository>()
        val bookQueryService = mockk<BookQueryService>()
        val usecase = ReadAuthorBookUsecase(authorRepository, bookQueryService)
        val authorId = ID<Author>(1)
        val author = Author(
            id = authorId,
            name = "夏目漱石",
            birthDate = AuthorBirthDate(LocalDate.of(1867, 2, 9))
        )
        val expectedBooks = listOf(
            BookSummaryDto(
                id = ID(1),
                title = "日本文学選集",
                price = BookPrice.of(BigDecimal("3000.00")),
                status = BookPublishStatus.PUBLISHED
            )
        )

        every { authorRepository.findById(authorId) } returns author
        every { bookQueryService.findByAuthorId(authorId) } returns expectedBooks

        val result = usecase.execute(authorId)

        result shouldBe expectedBooks
        verify { authorRepository.findById(authorId) }
        verify { bookQueryService.findByAuthorId(authorId) }
    }
})
