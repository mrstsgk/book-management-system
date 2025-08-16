package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonValue

/**
 * 出版ステータス
 */
enum class BookStatus(@JsonValue val value: kotlin.Int) {

    PUBLISHED(1),
    UNPUBLISHED(2);


    companion object {
        fun of(value: kotlin.Int): BookStatus {
            return entries.first { it.value == value }
        }
    }
}
