package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import org.springframework.stereotype.Service

@Service
class CreateBookUsecase(
    private val repository: BookRepository,
    private val detailQueryService: BookDetailQueryService,
) {
    /**
     * 書籍を登録する
     */
    fun execute(command: CreateBookCommand): BookDto {
        val book = repository.insert(toEntity(command))

        return detailQueryService.findById(book.id!!)!!
    }

    private fun toEntity(command: CreateBookCommand) = Book(
        title = command.title,
        price = command.price,
        authorIds = command.authorIds,
        status = command.status,
    )
}
