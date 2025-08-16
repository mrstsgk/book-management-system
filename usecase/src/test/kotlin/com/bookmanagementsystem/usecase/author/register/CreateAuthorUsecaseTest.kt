package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.time.LocalDate

class CreateAuthorUsecaseTest : FunSpec({
    test("著者登録が正常に実行されること") {
        val repository = mockk<AuthorRepository>()
        val usecase = CreateAuthorUsecase(repository)
        val command = CreateAuthorCommand(
            name = "テスト著者",
            birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1))
        )
        val savedAuthor = Author(
            id = ID(1),
            name = "テスト著者",
            birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1))
        )

        every { repository.insert(any()) } returns savedAuthor

        val result = usecase.execute(command)
        result shouldBe AuthorDto(
            id = ID(1),
            name = "テスト著者",
            birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1))
        )
        verify {
            repository.insert(
                Author(
                    name = "テスト著者",
                    birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1))
                )
            )
        }
    }

    test("生年月日がnullの著者登録が正常に実行されること") {
        val repository = mockk<AuthorRepository>()
        val usecase = CreateAuthorUsecase(repository)
        val command = CreateAuthorCommand(
            name = "テスト著者2",
            birthDate = null
        )
        val savedAuthor = Author(
            id = ID(2),
            name = "テスト著者2",
            birthDate = null
        )

        every { repository.insert(any()) } returns savedAuthor

        val result = usecase.execute(command)
        result shouldBe AuthorDto(
            id = ID(2),
            name = "テスト著者2",
            birthDate = null
        )
        verify {
            repository.insert(
                Author(
                    name = "テスト著者2",
                    birthDate = null
                )
            )
        }
    }
})
