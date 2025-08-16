package com.bookmanagementsystem.domain.core

import java.math.BigDecimal

/**
 * 金額を扱う値オブジェクト
 */
@JvmInline
value class Amount(val value: BigDecimal) {
    companion object {
        fun of(amount: Long) = Amount(BigDecimal.valueOf(amount))
    }

    fun toLong() = value.toLong()
}
