package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.read.AuthorQueryService
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import com.bookmanagementsystem.usecase.exception.UsecaseViolationException
import com.bookmanagementsystem.usecase.validation.CommandValidator
import org.springframework.stereotype.Service

@Service
class UpdateBookUsecase(
    private val repository: BookRepository,
    private val detailQueryService: BookDetailQueryService,
    private val authorQueryService: AuthorQueryService,
    private val validator: CommandValidator,
) {
    /**
     * 書籍を更新する
     */
    fun execute(command: UpdateBookCommand): BookDto {
        detailQueryService.findById(command.id) ?: throw NoSuchElementException("書籍が存在しません: ${command.id}")

        val validationErrors = validator.validate(command)
        if (validationErrors.isNotEmpty()) throw UsecaseViolationException(validationErrors.joinToString(", "))

        val book = repository.update(toEntity(command))
        val authorDtoList = authorQueryService.findByBookId(book.id!!)

        return toDto(book, authorDtoList)
    }

    private fun toEntity(command: UpdateBookCommand) = Book(
        id = command.id,
        title = command.title,
        price = command.price,
        authorIds = command.authorIds,
        status = command.status,
        version = command.version
    )

    private fun toDto(book: Book, authorDtoList: List<AuthorDto>) = BookDto(
        id = book.id!!,
        title = book.title,
        price = book.price,
        authors = authorDtoList,
        status = book.status,
        version = book.version!!
    )
}
