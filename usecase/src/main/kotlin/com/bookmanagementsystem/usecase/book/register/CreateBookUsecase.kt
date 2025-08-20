package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import com.bookmanagementsystem.usecase.exception.UsecaseViolationException
import com.bookmanagementsystem.usecase.validation.CommandValidator
import org.springframework.stereotype.Service

@Service
class CreateBookUsecase(
    private val repository: BookRepository,
    private val validator: CommandValidator,
    private val detailQueryService: BookDetailQueryService,
) {
    /**
     * 書籍を登録する
     */
    fun execute(command: CreateBookCommand): BookDto {
        val validationErrors = validator.validate(command)
        if (validationErrors.isNotEmpty()) throw UsecaseViolationException(validationErrors)

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
