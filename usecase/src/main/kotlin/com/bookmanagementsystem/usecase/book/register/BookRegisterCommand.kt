package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorExists
import com.bookmanagementsystem.usecase.validation.core.UniqueId

sealed interface BookRegisterCommand

/**
 * 書籍を登録するコマンド
 */
data class CreateBookCommand(
    val title: String,
    val price: BookPrice,
    @UniqueId
    @AuthorExists
    val authorIds: List<ID<Author>>,
    val status: BookPublishStatus,
) : BookRegisterCommand

/**
 * 書籍を更新するコマンド
 */
@CanChangeBookPublishStatus
data class UpdateBookCommand(
    val id: ID<Book>,
    val title: String,
    val price: BookPrice,
    @UniqueId
    @AuthorExists
    val authorIds: List<ID<Author>>,
    val status: BookPublishStatus,
    val version: Int,
) : BookRegisterCommand
