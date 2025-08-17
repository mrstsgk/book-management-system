package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID

/**
 * 書籍を表現するエンティティ
 */
data class Book(
    val id: ID<Book>? = null,
    val title: String,
    val price: BookPrice,
    val authorIds: List<ID<Author>>,
    val status: BookPublishStatus,
) {
    init {
        require(authorIds.isNotEmpty())
    }
}
