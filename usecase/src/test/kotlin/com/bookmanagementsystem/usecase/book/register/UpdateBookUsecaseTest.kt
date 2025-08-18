package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.read.AuthorQueryService
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
    val detailQueryService = mockk<BookDetailQueryService>()
    val authorQueryService = mockk<AuthorQueryService>()
    val validator = mockk<CommandValidator>()
    val usecase = UpdateBookUsecase(repository, detailQueryService, authorQueryService, validator)

    beforeEach {
        clearMocks(repository, detailQueryService, authorQueryService, validator)
    }

    test("書籍更新が正常に実行されること") {
        val bookId = ID<Book>(1)
        val authorId = ID<Author>(1)
        val command = UpdateBookCommand(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authorIds = listOf(authorId),
            status = BookPublishStatus.UNPUBLISHED
        )
        val updatedBook = Book(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authorIds = listOf(authorId),
            status = BookPublishStatus.UNPUBLISHED
        )
        val authorDto = AuthorDto(
            id = authorId,
            name = "著者名",
            birthDate = null,
            version = 1
        )

        every { detailQueryService.findById(bookId) } returns BookDto(
            id = bookId,
            title = "元のタイトル",
            price = BookPrice.of(1500L),
            authors = listOf(authorDto),
            status = BookPublishStatus.PUBLISHED
        )
        every { validator.validate(command) } returns emptyList()
        every { repository.update(any<Book>()) } returns updatedBook
        every { authorQueryService.findByBookId(bookId) } returns listOf(authorDto)

        val result = usecase.execute(command)
        result shouldBe BookDto(
            id = bookId,
            title = "更新後のタイトル",
            price = BookPrice.of(2000L),
            authors = listOf(authorDto),
            status = BookPublishStatus.UNPUBLISHED
        )
        verify {
            detailQueryService.findById(bookId)
            validator.validate(command)
            repository.update(
                Book(
                    id = bookId,
                    title = "更新後のタイトル",
                    price = BookPrice.of(2000L),
                    authorIds = listOf(authorId),
                    status = BookPublishStatus.UNPUBLISHED
                )
            )
            authorQueryService.findByBookId(bookId)
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
            status = BookPublishStatus.PUBLISHED
        )
        val updatedBook = Book(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authorIds = authorIds,
            status = BookPublishStatus.PUBLISHED
        )
        val authorDtoList = listOf(
            AuthorDto(ID(1), "著者1", null, 1),
            AuthorDto(ID(2), "著者2", null, 1),
            AuthorDto(ID(3), "著者3", null, 1)
        )

        every { detailQueryService.findById(bookId) } returns BookDto(
            id = bookId,
            title = "複数著者の書籍",
            price = BookPrice.of(2500L),
            authors = authorDtoList,
            status = BookPublishStatus.UNPUBLISHED
        )
        every { validator.validate(command) } returns emptyList()
        every { repository.update(any<Book>()) } returns updatedBook
        every { authorQueryService.findByBookId(bookId) } returns authorDtoList

        val result = usecase.execute(command)
        result shouldBe BookDto(
            id = bookId,
            title = "複数著者の書籍更新版",
            price = BookPrice.of(3000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED
        )
        verify {
            detailQueryService.findById(bookId)
            validator.validate(command)
            repository.update(any<Book>())
            authorQueryService.findByBookId(bookId)
        }
    }

    test("存在しない書籍を更新しようとするとNoSuchElementExceptionがスローされること") {
        val bookId = ID<Book>(999)
        val command = UpdateBookCommand(
            id = bookId,
            title = "存在しない書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID<Author>(1)),
            status = BookPublishStatus.PUBLISHED
        )

        every { detailQueryService.findById(bookId) } returns null

        val exception = shouldThrow<NoSuchElementException> {
            usecase.execute(command)
        }
        exception.message shouldBe "書籍が見つかりません: ID(value=999)"
        verify {
            detailQueryService.findById(bookId)
        }
        verify(exactly = 0) {
            validator.validate(any<UpdateBookCommand>())
            repository.update(any<Book>())
            authorQueryService.findByBookId(any<ID<Book>>())
        }
    }

    test("バリデーションエラーがある場合UsecaseViolationExceptionがスローされること") {
        val bookId = ID<Book>(1)
        val command = UpdateBookCommand(
            id = bookId,
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID<Author>(1)),
            status = BookPublishStatus.UNPUBLISHED
        )
        val validationErrors = listOf("status: 出版状況は「出版済み」から「未出版」に変更できません")

        every { detailQueryService.findById(bookId) } returns BookDto(
            id = bookId,
            title = "既存書籍",
            price = BookPrice.of(1500L),
            authors = emptyList(),
            status = BookPublishStatus.PUBLISHED
        )
        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.message shouldBe "status: 出版状況は「出版済み」から「未出版」に変更できません"
        verify {
            detailQueryService.findById(bookId)
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.update(any<Book>())
            authorQueryService.findByBookId(any<ID<Book>>())
        }
    }
})
