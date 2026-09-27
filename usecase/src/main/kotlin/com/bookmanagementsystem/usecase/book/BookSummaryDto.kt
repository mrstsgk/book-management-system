package com.bookmanagementsystem.usecase.book

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID

/**
 * 書籍の要約情報を表すDTO
 */
data class BookSummaryDto(
    val id: ID<Book>,
    val title: String,
    val price: BookPrice,
    val status: BookPublishStatus,
)
