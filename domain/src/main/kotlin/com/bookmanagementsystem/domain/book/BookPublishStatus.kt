package com.bookmanagementsystem.domain.book

/**
 * 出版ステータスを表現する列挙型
 */
enum class BookPublishStatus(val value: Int) {
    UNPUBLISHED(1), // 未出版
    PUBLISHED(2), // 出版済み
    ;

    companion object {
        fun of(value: Int) = entries.first { it.value == value }
    }

    /**
     * 出版可能かどうかを判定する
     */
    fun canPublish() = when (this) {
        UNPUBLISHED -> true
        PUBLISHED -> false
    }
}
