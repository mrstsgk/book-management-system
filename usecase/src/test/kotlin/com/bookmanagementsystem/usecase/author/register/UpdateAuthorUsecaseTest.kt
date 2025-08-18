package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.time.LocalDate

class UpdateAuthorUsecaseTest : FunSpec({
    test("著者更新が正常に実行されること") {
        val repository = mockk<AuthorRepository>()
        val usecase = UpdateAuthorUsecase(repository)
        val authorId = ID<Author>(1)
        val command = UpdateAuthorCommand(
            id = authorId,
            name = "更新後の著者名",
            birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
            version = 1
        )
        val existingAuthor = Author(
            id = authorId,
            name = "元の著者名",
            birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
            version = 1
        )
        val updatedAuthor = Author(
            id = authorId,
            name = "更新後の著者名",
            birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
            version = 1
        )

        every { repository.findById(authorId) } returns existingAuthor
        every { repository.update(any()) } returns updatedAuthor

        val result = usecase.execute(command)
        result shouldBe AuthorDto(
            id = authorId,
            name = "更新後の著者名",
            birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
            version = 1
        )
        verify {
            repository.findById(authorId)
            repository.update(
                Author(
                    id = authorId,
                    name = "更新後の著者名",
                    birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
                    version = 1
                )
            )
        }
    }

    test("生年月日をnullに更新できること") {
        val repository = mockk<AuthorRepository>()
        val usecase = UpdateAuthorUsecase(repository)
        val authorId = ID<Author>(2)
        val command = UpdateAuthorCommand(
            id = authorId,
            name = "更新後の著者名2",
            birthDate = null,
            version = 1
        )
        val existingAuthor = Author(
            id = authorId,
            name = "元の著者名2",
            birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
            version = 1
        )
        val updatedAuthor = Author(
            id = authorId,
            name = "更新後の著者名2",
            birthDate = null,
            version = 1
        )

        every { repository.findById(authorId) } returns existingAuthor
        every { repository.update(any()) } returns updatedAuthor

        val result = usecase.execute(command)
        result shouldBe AuthorDto(
            id = authorId,
            name = "更新後の著者名2",
            birthDate = null,
            version = 1
        )
        verify {
            repository.findById(authorId)
            repository.update(
                Author(
                    id = authorId,
                    name = "更新後の著者名2",
                    birthDate = null,
                    version = 1
                )
            )
        }
    }

    test("存在しない著者を更新しようとするとNoSuchElementExceptionがスローされること") {
        val repository = mockk<AuthorRepository>()
        val usecase = UpdateAuthorUsecase(repository)
        val authorId = ID<Author>(999)
        val command = UpdateAuthorCommand(
            id = authorId,
            name = "更新後の著者名",
            birthDate = AuthorBirthDate(LocalDate.of(1985, 5, 15)),
            version = 1
        )

        every { repository.findById(authorId) } returns null

        val exception = shouldThrow<NoSuchElementException> {
            usecase.execute(command)
        }
        exception.message shouldBe "著者が見つかりません: ID(value=999)"
        verify {
            repository.findById(authorId)
        }
        verify(exactly = 0) {
            repository.update(any())
        }
    }
})
