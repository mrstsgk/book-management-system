package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonValue

/**
 * 出版ステータス(1：未出版、2：出版済み)
 */
enum class BookStatus(@JsonValue val value: kotlin.Int) {

    UNPUBLISHED(1),
    PUBLISHED(2);


    companion object {
        fun of(value: kotlin.Int): BookStatus {
            return entries.first { it.value == value }
        }
    }
}
