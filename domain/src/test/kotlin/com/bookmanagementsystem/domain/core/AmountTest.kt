package com.bookmanagementsystem.domain.core

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import java.math.BigDecimal

class AmountTest : FunSpec({
    context("金額の生成") {
        test("Long値から金額を生成できる") {
            val amount = Amount.of(1000L)
            amount.value shouldBe BigDecimal.valueOf(1000)
        }

        test("BigDecimalから金額を生成できる") {
            val amount = Amount(BigDecimal("1000.50"))
            amount.value shouldBe BigDecimal("1000.50")
        }

        test("負の金額も扱える") {
            val amount = Amount.of(-500L)
            amount.value shouldBe BigDecimal.valueOf(-500)
        }

        test("ゼロの金額も扱える") {
            val amount = Amount.of(0L)
            amount.value shouldBe BigDecimal.ZERO
        }
    }

    context("金額の変換") {
        test("金額をLong値に変換できる") {
            val amount = Amount.of(1000L)
            amount.toLong() shouldBe 1000L
        }
    }
})