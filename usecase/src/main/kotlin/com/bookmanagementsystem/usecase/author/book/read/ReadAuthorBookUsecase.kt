package com.bookmanagementsystem.usecase.author.book.read

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.usecase.book.BookSummaryDto
import com.bookmanagementsystem.usecase.book.read.BookQueryService
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional

@Service
class ReadAuthorBookUsecase(
    private val authorRepository: AuthorRepository,
    private val bookQueryService: BookQueryService
) {
    /**
     * 指定された著者IDに紐づく書籍一覧を取得する
     */
    @Transactional(readOnly = true)
    fun execute(authorId: ID<Author>): List<BookSummaryDto> {
        authorRepository.findById(authorId)
            ?: throw NoSuchElementException("著者が見つかりません。ID: ${authorId.value}")

        return bookQueryService.findByAuthorId(authorId)
    }
}
