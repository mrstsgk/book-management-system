package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import com.bookmanagementsystem.usecase.exception.UsecaseViolationException
import com.bookmanagementsystem.usecase.validation.CommandValidator
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.clearMocks
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify

class UpdateBookUsecaseTest : FunSpec({
    val repository = mockk<BookRepository>()
    val validator = mockk<CommandValidator>()
    val detailQueryService = mockk<BookDetailQueryService>()
    val usecase = UpdateBookUsecase(repository, validator, detailQueryService)

    beforeEach {
        clearMocks(repository, validator, detailQueryService)
    }

    test("書籍更新が正常に実行されること") {
        val bookId = ID<Book>(1)
        val authorId = ID<Author>(1)
        val command = UpdateBookCommand(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authorIds = listOf(authorId),
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )
        val updatedBook = Book(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authorIds = listOf(authorId),
            status = BookPublishStatus.UNPUBLISHED,
            version = 2
        )
        val authorDto = AuthorDto(
            id = authorId,
            name = "著者名",
            birthDate = null,
            version = 1
        )

        every { repository.findById(bookId) } returns Book(
            id = bookId,
            title = "元のタイトル",
            price = BookPrice.of(1500L),
            authorIds = listOf(authorId),
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )
        every { validator.validate(command) } returns emptyList()
        every { repository.update(any<Book>()) } returns updatedBook
        every { detailQueryService.findById(bookId) } returns BookDto(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authors = listOf(authorDto),
            status = BookPublishStatus.UNPUBLISHED,
            version = 2
        )

        val result = usecase.execute(command)
        result shouldBe BookDto(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authors = listOf(authorDto),
            status = BookPublishStatus.UNPUBLISHED,
            version = 2
        )
        verify {
            repository.findById(bookId)
            validator.validate(command)
            repository.update(
                Book(
                    id = bookId,
                    title = "更新後のタイトル",
                    price = BookPrice.of(2000L),
                    authorIds = listOf(authorId),
                    status = BookPublishStatus.UNPUBLISHED,
                    version = 1
                )
            )
            detailQueryService.findById(bookId)
        }
    }

    test("複数の著者を持つ書籍を更新できること") {
        val bookId = ID<Book>(2)
        val authorIds: List<ID<Author>> = listOf(ID(1), ID(2), ID(3))
        val command = UpdateBookCommand(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authorIds = authorIds,
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )
        val updatedBook = Book(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authorIds = authorIds,
            status = BookPublishStatus.PUBLISHED,
            version = 2
        )
        val authorDtoList = listOf(
            AuthorDto(ID(1), "著者1", null, 1),
            AuthorDto(ID(2), "著者2", null, 1),
            AuthorDto(ID(3), "著者3", null, 1)
        )

        every { repository.findById(bookId) } returns Book(
            id = bookId,
            title = "複数著者の書籍",
            price = BookPrice.of(2500L),
            authorIds = authorIds,
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )
        every { validator.validate(command) } returns emptyList()
        every { repository.update(any<Book>()) } returns updatedBook
        every { detailQueryService.findById(bookId) } returns BookDto(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 2
        )

        val result = usecase.execute(command)
        result shouldBe BookDto(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 2
        )
        verify {
            repository.findById(bookId)
            validator.validate(command)
            repository.update(any<Book>())
            detailQueryService.findById(bookId)
        }
    }

    test("バリデーションエラーがある場合UsecaseViolationExceptionがスローされること") {
        val bookId = ID<Book>(1)
        val command = UpdateBookCommand(
            id = bookId,
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1)),
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )
        val validationErrors = listOf("status: 出版状況は「出版済み」から「未出版」に変更できません")

        every { repository.findById(bookId) } returns Book(
            id = bookId,
            title = "既存書籍",
            price = BookPrice.of(1500L),
            authorIds = listOf(ID(1)),
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )
        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.message shouldBe "status: 出版状況は「出版済み」から「未出版」に変更できません"
        verify {
            repository.findById(bookId)
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.update(any<Book>())
            detailQueryService.findById(any<ID<Book>>())
        }
    }
})
