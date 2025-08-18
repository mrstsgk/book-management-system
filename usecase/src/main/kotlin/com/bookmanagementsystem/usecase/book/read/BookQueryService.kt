package com.bookmanagementsystem.usecase.book.read

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.book.BookDto

interface BookQueryService {
    /**
     * 著者IDに紐づく書籍一覧を取得する
     */
    fun findByAuthorId(authorId: ID<Author>): List<BookDto>
}
