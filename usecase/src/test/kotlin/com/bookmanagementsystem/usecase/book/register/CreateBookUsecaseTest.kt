package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.author.AuthorBirthDate
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
import java.time.LocalDate

class CreateBookUsecaseTest : FunSpec({
    val repository = mockk<BookRepository>()
    val validator = mockk<CommandValidator>()
    val detailQueryService = mockk<BookDetailQueryService>()
    val usecase = CreateBookUsecase(repository, validator, detailQueryService)

    beforeEach {
        clearMocks(repository, validator, detailQueryService)
    }

    test("書籍登録が正常に実行されること") {

        val command = CreateBookCommand(
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(2)),
            status = BookPublishStatus.PUBLISHED
        )

        val savedBook = Book(
            id = ID(1),
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(2)),
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        val authorDtoList = listOf(
            AuthorDto(
                id = ID(1),
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
                version = 1
            ),
            AuthorDto(
                id = ID(2),
                name = "著者2",
                birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
                version = 1
            )
        )

        every { validator.validate(command) } returns emptyList()
        every { repository.insert(any()) } returns savedBook
        every { detailQueryService.findById(ID(1)) } returns BookDto(
            id = ID(1),
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        val result = usecase.execute(command)

        result shouldBe BookDto(
            id = ID(1),
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        verify {
            validator.validate(command)
            repository.insert(
                Book(
                    title = "テスト書籍",
                    price = BookPrice.of(1000L),
                    authorIds = listOf(ID(1), ID(2)),
                    status = BookPublishStatus.PUBLISHED
                )
            )
            detailQueryService.findById(ID(1))
        }
    }

    test("未出版ステータスの書籍登録が正常に実行されること") {

        val command = CreateBookCommand(
            title = "未出版書籍",
            price = BookPrice.of(1500L),
            authorIds = listOf(ID(1)),
            status = BookPublishStatus.UNPUBLISHED
        )

        val savedBook = Book(
            id = ID(2),
            title = "未出版書籍",
            price = BookPrice.of(1500L),
            authorIds = listOf(ID(1)),
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )

        val authorDtoList = listOf(
            AuthorDto(
                id = ID(1),
                name = "著者1",
                birthDate = null,
                version = 1
            )
        )

        every { validator.validate(command) } returns emptyList()
        every { repository.insert(any()) } returns savedBook
        every { detailQueryService.findById(ID(2)) } returns BookDto(
            id = ID(2),
            title = "未出版書籍",
            price = BookPrice.of(1500L),
            authors = authorDtoList,
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )

        val result = usecase.execute(command)

        result shouldBe BookDto(
            id = ID(2),
            title = "未出版書籍",
            price = BookPrice.of(1500L),
            authors = authorDtoList,
            status = BookPublishStatus.UNPUBLISHED,
            version = 1
        )

        verify {
            validator.validate(command)
            repository.insert(
                Book(
                    title = "未出版書籍",
                    price = BookPrice.of(1500L),
                    authorIds = listOf(ID(1)),
                    status = BookPublishStatus.UNPUBLISHED
                )
            )
            detailQueryService.findById(ID(2))
        }
    }

    test("複数著者の書籍登録が正常に実行されること") {

        val command = CreateBookCommand(
            title = "共著書籍",
            price = BookPrice.of(2000L),
            authorIds = listOf(ID(1), ID(2), ID(3)),
            status = BookPublishStatus.PUBLISHED
        )

        val savedBook = Book(
            id = ID(3),
            title = "共著書籍",
            price = BookPrice.of(2000L),
            authorIds = listOf(ID(1), ID(2), ID(3)),
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        val authorDtoList = listOf(
            AuthorDto(
                id = ID(1),
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1980, 3, 10)),
                version = 1
            ),
            AuthorDto(
                id = ID(2),
                name = "著者2",
                birthDate = AuthorBirthDate(LocalDate.of(1975, 8, 20)),
                version = 1
            ),
            AuthorDto(
                id = ID(3),
                name = "著者3",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 12, 5)),
                version = 1
            )
        )

        every { validator.validate(command) } returns emptyList()
        every { repository.insert(any()) } returns savedBook
        every { detailQueryService.findById(ID(3)) } returns BookDto(
            id = ID(3),
            title = "共著書籍",
            price = BookPrice.of(2000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        val result = usecase.execute(command)

        result shouldBe BookDto(
            id = ID(3),
            title = "共著書籍",
            price = BookPrice.of(2000L),
            authors = authorDtoList,
            status = BookPublishStatus.PUBLISHED,
            version = 1
        )

        verify {
            validator.validate(command)
            repository.insert(
                Book(
                    title = "共著書籍",
                    price = BookPrice.of(2000L),
                    authorIds = listOf(ID(1), ID(2), ID(3)),
                    status = BookPublishStatus.PUBLISHED
                )
            )
            detailQueryService.findById(ID(3))
        }
    }

    test("重複した著者IDがある場合UsecaseViolationExceptionがスローされること") {
        val command = CreateBookCommand(
            title = "重複著者テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(2), ID(1)), // 著者ID 1が重複
            status = BookPublishStatus.PUBLISHED
        )
        val validationErrors = listOf("authorIds: リスト内に重複した要素があります")

        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.errors shouldBe listOf("authorIds: リスト内に重複した要素があります")
        verify {
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.insert(any<Book>())
            detailQueryService.findById(any<ID<Book>>())
        }
    }

    test("複数のバリデーションエラーがある場合UsecaseViolationExceptionがスローされること") {
        val command = CreateBookCommand(
            title = "", // 空のタイトル（Pattern違反）
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(2), ID(1)), // 著者ID 1が重複
            status = BookPublishStatus.PUBLISHED
        )
        val validationErrors = listOf(
            "title: タイトルは空白文字のみで構成することはできません",
            "authorIds: リスト内に重複した要素があります"
        )

        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.errors shouldBe listOf(
            "title: タイトルは空白文字のみで構成することはできません",
            "authorIds: リスト内に重複した要素があります"
        )
        verify {
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.insert(any<Book>())
            detailQueryService.findById(any<ID<Book>>())
        }
    }

    test("存在しない著者IDが指定された場合UsecaseViolationExceptionがスローされること") {
        val command = CreateBookCommand(
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(999)), // ID 999は存在しない
            status = BookPublishStatus.PUBLISHED
        )
        val validationErrors = listOf("authorIds: 存在しない著者が指定されています")

        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.errors shouldBe listOf("authorIds: 存在しない著者が指定されています")
        verify {
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.insert(any<Book>())
            detailQueryService.findById(any<ID<Book>>())
        }
    }

    test("重複した著者IDと存在しない著者IDの両方のエラーが発生した場合UsecaseViolationExceptionがスローされること") {
        val command = CreateBookCommand(
            title = "テスト書籍",
            price = BookPrice.of(1000L),
            authorIds = listOf(ID(1), ID(1), ID(999)), // ID重複と存在しないID
            status = BookPublishStatus.PUBLISHED
        )
        val validationErrors = listOf(
            "authorIds: リスト内に重複した要素があります",
            "authorIds: 存在しない著者が指定されています"
        )

        every { validator.validate(command) } returns validationErrors

        val exception = shouldThrow<UsecaseViolationException> {
            usecase.execute(command)
        }
        exception.errors shouldBe listOf(
            "authorIds: リスト内に重複した要素があります",
            "authorIds: 存在しない著者が指定されています"
        )
        verify {
            validator.validate(command)
        }
        verify(exactly = 0) {
            repository.insert(any<Book>())
            detailQueryService.findById(any<ID<Book>>())
        }
    }
})
