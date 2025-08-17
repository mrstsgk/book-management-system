package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.core.Amount
import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import java.math.BigDecimal

class BookPriceTest : FunSpec({
    context("書籍価格の生成") {
        test("正の金額で書籍価格を作成できる") {
            val price = BookPrice.of(1000L)
            price.toLong() shouldBe 1000L
        }

        test("ゼロ円で書籍価格を作成できる") {
            val price = BookPrice.of(0L)
            price.toLong() shouldBe 0L
        }

        test("負の金額で書籍価格を作成できない") {
            shouldThrow<IllegalArgumentException> {
                BookPrice.of(-100L)
            }
        }

        test("Amountから書籍価格を作成できる") {
            val amount = Amount(BigDecimal("1500.50"))
            val price = BookPrice(amount)
            price.amount shouldBe amount
        }

        test("小数を含む書籍金額で価格を作成できる") {
            val price = BookPrice.of(BigDecimal("1500.50"))
            price.amount.value shouldBe BigDecimal("1500.50")
        }
    }

    context("書籍価格の変換") {
        test("価格をLong値に変換できる") {
            val price = BookPrice.of(2500L)
            price.toLong() shouldBe 2500L
        }
    }
})
