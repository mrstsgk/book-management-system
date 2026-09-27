package com.bookmanagementsystem.usecase.author.register

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import java.time.LocalDate

/**
 * PastOnlyValidatorのテスト
 */
class PastOnlyValidatorTest : FunSpec({
    val validator = PastOnlyValidator()

    context("過去日付のバリデーション") {
        test("nullの場合はtrue") {
            val result = validator.isValid(null, null)

            result shouldBe true
        }

        test("過去の日付の場合はtrue") {
            val pastDate = LocalDate.now().minusDays(1)
            val result = validator.isValid(pastDate, null)

            result shouldBe true
        }

        test("今日の日付の場合はfalse") {
            val today = LocalDate.now()
            val result = validator.isValid(today, null)

            result shouldBe false
        }

        test("未来の日付の場合はfalse") {
            val futureDate = LocalDate.now().plusDays(1)
            val result = validator.isValid(futureDate, null)

            result shouldBe false
        }

        test("1年前の日付の場合はtrue") {
            val oneYearAgo = LocalDate.now().minusYears(1)
            val result = validator.isValid(oneYearAgo, null)

            result shouldBe true
        }

        test("1年後の日付の場合はfalse") {
            val oneYearLater = LocalDate.now().plusYears(1)
            val result = validator.isValid(oneYearLater, null)

            result shouldBe false
        }
    }
})
