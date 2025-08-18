package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK_AUTHOR
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import org.jooq.DSLContext
import org.jooq.Record
import org.springframework.stereotype.Repository

@Repository
class BookDetailQueryServiceImpl(private val dsl: DSLContext) : BookDetailQueryService {
    override fun findById(id: ID<Book>): BookDto? {
        val records = dsl.select(
            BOOK.ID,
            BOOK.TITLE,
            BOOK.PRICE,
            BOOK.PUBLISH_STATUS,
            AUTHOR.ID.`as`("author_id"),
            AUTHOR.NAME.`as`("author_name"),
            AUTHOR.BIRTH_DATE.`as`("author_birth_date")
        )
            .from(BOOK)
            .leftJoin(BOOK_AUTHOR).on(BOOK.ID.eq(BOOK_AUTHOR.BOOK_ID))
            .leftJoin(AUTHOR).on(BOOK_AUTHOR.AUTHOR_ID.eq(AUTHOR.ID))
            .where(BOOK.ID.eq(id.value))
            .fetch()

        if (records.isEmpty()) return null

        val bookRecord = records.first()
        val authors = convertAuthorDtoList(records)

        return convert(bookRecord, authors)
    }

    private fun convertAuthorDtoList(records: List<Record>): List<AuthorDto> {
        return records.mapNotNull { record ->
            val authorId = record.getValue("author_id") as? Int ?: return@mapNotNull null
            val authorName = record.getValue("author_name") as? String ?: return@mapNotNull null
            val authorBirthDate = record.getValue("author_birth_date") as? java.time.LocalDate

            AuthorDto(
                id = ID(authorId),
                name = authorName,
                birthDate = authorBirthDate?.let { AuthorBirthDate(it) }
            )
        }
    }

    private fun convert(bookRecord: Record, authors: List<AuthorDto>): BookDto {
        return BookDto(
            id = ID(bookRecord[BOOK.ID]!!),
            title = bookRecord[BOOK.TITLE]!!,
            price = BookPrice.of(bookRecord[BOOK.PRICE]!!),
            authors = authors,
            status = BookPublishStatus.of(bookRecord[BOOK.PUBLISH_STATUS]!!)
        )
    }
}
