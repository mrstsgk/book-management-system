package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.core.Amount
import java.math.BigDecimal

/**
 * 書籍価格を表現する値オブジェクト
 */
data class BookPrice(val amount: Amount) {
    init {
        require(BigDecimal.ZERO <= amount.value)
    }

    companion object {
        fun of(amount: Long): BookPrice = BookPrice(Amount.of(amount))
    }

    fun toLong(): Long = amount.toLong()
}
