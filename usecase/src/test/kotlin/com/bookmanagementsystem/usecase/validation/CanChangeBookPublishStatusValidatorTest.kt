package com.bookmanagementsystem.usecase.validation

import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.book.register.UpdateBookCommand
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import io.mockk.every
import io.mockk.mockk

/**
 * CanChangeBookPublishStatusValidatorのテスト
 */
class CanChangeBookPublishStatusValidatorTest : FunSpec({
    val repository = mockk<BookRepository>()
    val validator = CanChangeBookPublishStatusValidator(repository)

    context("出版ステータス変更のバリデーション") {
        test("UNPUBLISHEDからUNPUBLISHEDへの変更は有効") {
            val command = UpdateBookCommand(
                id = ID(1),
                title = "テスト書籍",
                price = BookPrice.of(1000L),
                authorIds = listOf(ID(1)),
                status = BookPublishStatus.UNPUBLISHED
            )
            every { repository.getBookPublishStatusById(ID(1)) } returns BookPublishStatus.UNPUBLISHED

            val result = validator.isValid(command, null)

            result shouldBe true
        }

        test("UNPUBLISHEDからPUBLISHEDへの変更は有効") {
            val command = UpdateBookCommand(
                id = ID(1),
                title = "テスト書籍",
                price = BookPrice.of(1000L),
                authorIds = listOf(ID(1)),
                status = BookPublishStatus.PUBLISHED
            )
            every { repository.getBookPublishStatusById(ID(1)) } returns BookPublishStatus.UNPUBLISHED

            val result = validator.isValid(command, null)

            result shouldBe true
        }

        test("PUBLISHEDからPUBLISHEDへの変更は有効") {
            val command = UpdateBookCommand(
                id = ID(1),
                title = "テスト書籍",
                price = BookPrice.of(1000L),
                authorIds = listOf(ID(1)),
                status = BookPublishStatus.PUBLISHED
            )
            every { repository.getBookPublishStatusById(ID(1)) } returns BookPublishStatus.PUBLISHED

            val result = validator.isValid(command, null)

            result shouldBe true
        }

        test("PUBLISHEDからUNPUBLISHEDへの変更は無効") {
            val command = UpdateBookCommand(
                id = ID(1),
                title = "テスト書籍",
                price = BookPrice.of(1000L),
                authorIds = listOf(ID(1)),
                status = BookPublishStatus.UNPUBLISHED
            )
            every { repository.getBookPublishStatusById(ID(1)) } returns BookPublishStatus.PUBLISHED

            val result = validator.isValid(command, null)

            result shouldBe false
        }
    }
})
