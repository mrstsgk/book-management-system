package com.bookmanagementsystem.domain.author

import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import java.time.LocalDate

class AuthorBirthDateTest : FunSpec({
    test("過去の日付で誕生日を作成できる") {
        val pastDate = LocalDate.of(1990, 5, 15)
        val authorBirthDate = AuthorBirthDate(pastDate)
        authorBirthDate.value shouldBe pastDate
    }

    test("昨日の日付で誕生日を作成できる") {
        val yesterday = LocalDate.now().minusDays(1)
        val authorBirthDate = AuthorBirthDate(yesterday)
        authorBirthDate.value shouldBe yesterday
    }

    test("現在日付と等しいときは誕生日を作成できない") {
        val today = LocalDate.now()
        shouldThrow<IllegalArgumentException> {
            AuthorBirthDate(today)
        }
    }

    test("現在日付よりも未来の日付で誕生日を作成できない") {
        val futureDate = LocalDate.now().plusDays(1)
        shouldThrow<IllegalArgumentException> {
            AuthorBirthDate(futureDate)
        }
    }
})
