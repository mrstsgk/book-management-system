package com.bookmanagementsystem.usecase.book

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto

/**
 * 書籍のDTO
 */
data class BookDto(
    val id: ID<Book>,
    val title: String,
    val price: BookPrice,
    val authors: List<AuthorDto>,
    val status: BookPublishStatus,
)
