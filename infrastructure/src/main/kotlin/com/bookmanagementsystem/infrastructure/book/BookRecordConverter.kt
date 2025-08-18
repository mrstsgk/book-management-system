package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.book.BookDto
import org.jooq.Record

object BookRecordConverter {
    /**
     * 著者情報を含む書籍レコードのリストから、著者DTOのリストを変換する
     */
    fun convertAuthorDtoList(records: List<Record>): List<AuthorDto> {
        return records.map { record ->
            val authorId = record.get("author_id", Int::class.java)!!
            val authorName = record.get("author_name", String::class.java)!!
            val authorBirthDate = record.get("author_birth_date", java.time.LocalDate::class.java)
            val authorVersion = record.get("author_version", Int::class.java)!!

            AuthorDto(
                id = ID(authorId),
                name = authorName,
                birthDate = authorBirthDate?.let { AuthorBirthDate(it) },
                version = authorVersion
            )
        }
    }

    /**
     * 書籍レコードと著者DTOのリストから、書籍DTOを変換する
     */
    fun convert(bookRecord: Record, authors: List<AuthorDto>): BookDto {
        return BookDto(
            id = ID(bookRecord[BOOK.ID]!!),
            title = bookRecord[BOOK.TITLE]!!,
            price = BookPrice.of(bookRecord[BOOK.PRICE]!!),
            authors = authors,
            status = BookPublishStatus.of(bookRecord[BOOK.PUBLISH_STATUS]!!),
        )
    }
}
