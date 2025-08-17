package com.bookmanagementsystem.usecase.author.read

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.author.AuthorDto

interface AuthorQueryService {
    /**
     * 書籍に紐づいている著者を取得する
     */
    fun findByBookId(bookId: ID<Book>): List<AuthorDto>
}
