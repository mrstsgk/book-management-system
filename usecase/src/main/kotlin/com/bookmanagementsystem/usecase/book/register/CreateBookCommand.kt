package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID

/**
 * 書籍を登録するコマンド
 */
data class CreateBookCommand(
    val title: String,
    val price: BookPrice,
    val authorIds: List<ID<Author>>,
    val status: BookPublishStatus,
)
