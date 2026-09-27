package com.bookmanagementsystem.domain.author

import java.time.LocalDate

/**
 * 生年月日を表現する値オブジェクト
 */
@JvmInline
value class AuthorBirthDate(val value: LocalDate) {
    init {
        require(value.isBefore(LocalDate.now()))
    }
}
