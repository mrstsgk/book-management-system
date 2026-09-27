package com.bookmanagementsystem.usecase.author

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.clearMocks
import io.mockk.every
import io.mockk.mockk
import java.time.LocalDate

/**
 * AuthorExistsValidatorのテスト
 */
class AuthorExistsValidatorTest : FunSpec({
    val repository = mockk<AuthorRepository>()
    val validator = AuthorExistsValidator(repository)

    beforeEach {
        clearMocks(repository)
    }

    context("著者存在チェックのバリデーション") {
        test("空のリストの場合はtrue") {
            val result = validator.isValid(emptyList(), null)

            result shouldBe true
        }

        test("すべての著者IDが存在する場合はtrue") {
            val authorId1 = ID<Author>(1)
            val authorId2 = ID<Author>(2)
            val authorIds = listOf(authorId1, authorId2)

            val author1 = Author(
                id = authorId1,
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
                version = 1
            )
            val author2 = Author(
                id = authorId2,
                name = "著者2",
                birthDate = AuthorBirthDate(LocalDate.of(1991, 2, 2)),
                version = 1
            )

            every { repository.findByIds(authorIds) } returns listOf(author1, author2)

            val result = validator.isValid(authorIds, null)

            result shouldBe true
        }

        test("存在しない著者IDが含まれる場合はfalse") {
            val authorId1 = ID<Author>(1)
            val authorId2 = ID<Author>(999) // 存在しないID
            val authorIds = listOf(authorId1, authorId2)

            val author1 = Author(
                id = authorId1,
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
                version = 1
            )

            every { repository.findByIds(authorIds) } returns listOf(author1) // author2は見つからない

            val result = validator.isValid(authorIds, null)

            result shouldBe false
        }

        test("重複したIDが含まれる場合でも正しく検証する") {
            val authorId1 = ID<Author>(1)
            val authorIds = listOf(authorId1, authorId1) // 重複

            val author1 = Author(
                id = authorId1,
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
                version = 1
            )

            every { repository.findByIds(listOf(authorId1)) } returns listOf(author1) // distinct()により重複は除去される

            val result = validator.isValid(authorIds, null)

            result shouldBe true
        }

        test("すべての著者IDが存在しない場合はfalse") {
            val authorId1 = ID<Author>(998)
            val authorId2 = ID<Author>(999)
            val authorIds = listOf(authorId1, authorId2)

            every { repository.findByIds(authorIds) } returns emptyList()

            val result = validator.isValid(authorIds, null)

            result shouldBe false
        }

        test("単一の存在する著者IDの場合はtrue") {
            val authorId1 = ID<Author>(1)
            val authorIds = listOf(authorId1)

            val author1 = Author(
                id = authorId1,
                name = "著者1",
                birthDate = AuthorBirthDate(LocalDate.of(1990, 1, 1)),
                version = 1
            )

            every { repository.findByIds(authorIds) } returns listOf(author1)

            val result = validator.isValid(authorIds, null)

            result shouldBe true
        }

        test("単一の存在しない著者IDの場合はfalse") {
            val authorId1 = ID<Author>(999)
            val authorIds = listOf(authorId1)

            every { repository.findByIds(authorIds) } returns emptyList()

            val result = validator.isValid(authorIds, null)

            result shouldBe false
        }
    }
})
