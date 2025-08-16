package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID

/**
 * 書籍を表現するエンティティ
 */
data class Book(
    val id: ID<Book>,
    val title: String,
    val price: BookPrice,
    val status: BookPublishStatus,
    val authorIds: List<ID<Author>>,
) {
    init {
        require(authorIds.isNotEmpty())
    }
}
