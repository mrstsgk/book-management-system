package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.read.AuthorQueryService
import com.bookmanagementsystem.usecase.book.BookDto
import org.springframework.stereotype.Service

@Service
class CreateBookUsecase(
    private val repository: BookRepository,
    private val authorQueryService: AuthorQueryService,
) {
    /**
     * 書籍を登録する
     */
    fun execute(command: CreateBookCommand): BookDto {
        val book = repository.insert(toEntity(command))
        val authorDtoList = authorQueryService.findByBookId(book.id!!)

        return toDto(book, authorDtoList)
    }

    private fun toEntity(command: CreateBookCommand) = Book(
        title = command.title,
        price = command.price,
        authorIds = command.authorIds,
        status = command.status,
    )

    private fun toDto(book: Book, authorDtoList: List<AuthorDto>) = BookDto(
        id = book.id!!,
        title = book.title,
        price = book.price,
        authors = authorDtoList,
        status = book.status,
        version = book.version!!,
    )
}
