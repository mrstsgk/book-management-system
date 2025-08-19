package com.bookmanagementsystem.usecase.book.read

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.book.BookDto

interface BookDetailQueryService {
    /**
     * IDに紐づく書籍の詳細情報を取得する
     */
    fun findById(id: ID<Book>): BookDto
}
