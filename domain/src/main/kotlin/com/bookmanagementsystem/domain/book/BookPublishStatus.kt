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
     * 指定されたステータスへの変更が可能かどうかを判定する
     */
    fun canChange(newStatus: BookPublishStatus): Boolean {
        return when (this) {
            UNPUBLISHED -> true // 未出版からはどの状態にも変更可能
            PUBLISHED -> newStatus == PUBLISHED // 出版済みからは出版済みのみ
        }
    }
}
