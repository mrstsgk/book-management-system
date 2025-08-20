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
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.time.LocalDate

class CreateBookUsecaseTest : FunSpec({
    test("書籍登録が正常に実行されること") {
        val repository = mockk<BookRepository>()
        val detailQueryService = mockk<BookDetailQueryService>()
        val usecase = CreateBookUsecase(repository, detailQueryService)

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
            repository.insert(
                Book(
                    title = "テスト書籍",
                    price = BookPrice.of(1000L),
                    authorIds = listOf(ID(1), ID(2)),
                    status = BookPublishStatus.PUBLISHED
                )
            )
        }
        verify {
            detailQueryService.findById(ID(1))
        }
    }

    test("未出版ステータスの書籍登録が正常に実行されること") {
        val repository = mockk<BookRepository>()
        val detailQueryService = mockk<BookDetailQueryService>()
        val usecase = CreateBookUsecase(repository, detailQueryService)

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
            repository.insert(
                Book(
                    title = "未出版書籍",
                    price = BookPrice.of(1500L),
                    authorIds = listOf(ID(1)),
                    status = BookPublishStatus.UNPUBLISHED
                )
            )
        }
    }

    test("複数著者の書籍登録が正常に実行されること") {
        val repository = mockk<BookRepository>()
        val detailQueryService = mockk<BookDetailQueryService>()
        val usecase = CreateBookUsecase(repository, detailQueryService)

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
            repository.insert(
                Book(
                    title = "共著書籍",
                    price = BookPrice.of(2000L),
                    authorIds = listOf(ID(1), ID(2), ID(3)),
                    status = BookPublishStatus.PUBLISHED
                )
            )
        }
    }
})
