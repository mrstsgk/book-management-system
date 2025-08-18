package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.read.AuthorQueryService
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import org.springframework.stereotype.Service

@Service
class UpdateBookUsecase(
    private val repository: BookRepository,
    private val detailQueryService: BookDetailQueryService,
    private val authorQueryService: AuthorQueryService,
) {
    /**
     * 書籍を更新する
     */
    fun execute(command: UpdateBookCommand): BookDto {
        detailQueryService.findById(command.id) ?: throw NoSuchElementException("書籍が見つかりません: ${command.id}")

        val book = repository.update(toEntity(command))
        val authorDtoList = authorQueryService.findByBookId(book.id!!)

        return toDto(book, authorDtoList)
    }

    private fun toEntity(command: UpdateBookCommand) = Book(
        id = command.id,
        title = command.title,
        price = command.price,
        authorIds = command.authorIds,
        status = command.status
    )

    private fun toDto(book: Book, authorDtoList: List<AuthorDto>) = BookDto(
        id = book.id!!,
        title = book.title,
        price = book.price,
        authors = authorDtoList,
        status = book.status
    )
}
