package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import com.bookmanagementsystem.usecase.exception.UsecaseViolationException
import com.bookmanagementsystem.usecase.validation.CommandValidator
import org.springframework.stereotype.Service

@Service
class UpdateBookUsecase(
    private val repository: BookRepository,
    private val validator: CommandValidator,
    private val queryService: BookDetailQueryService,
) {
    /**
     * 書籍を更新する
     */
    fun execute(command: UpdateBookCommand): BookDto {
        val currentBook =
            repository.findById(command.id) ?: throw NoSuchElementException("書籍が存在しません: ${command.id}")

        val validationErrors = validator.validate(command)
        if (validationErrors.isNotEmpty()) throw UsecaseViolationException(validationErrors.joinToString(", "))

        val book = repository.update(toEntity(currentBook, command))

        return queryService.findById(book.id!!)!!
    }

    private fun toEntity(book: Book, command: UpdateBookCommand) = book.copy(
        title = command.title,
        price = command.price,
        authorIds = command.authorIds,
        status = command.status,
        version = command.version,
    )
}
