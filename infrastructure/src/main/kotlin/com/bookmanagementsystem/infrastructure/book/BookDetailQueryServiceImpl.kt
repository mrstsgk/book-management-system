package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK
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
            AUTHOR.BIRTH_DATE.`as`("author_birth_date"),
            AUTHOR.VERSION.`as`("author_version"),
            BOOK.VERSION,
        )
            .from(BOOK)
            // NOTE: 書籍には必ず著者が存在するというビジネスルールが強制されるため innerJoin で実装する
            .innerJoin(AUTHOR_BOOK).on(BOOK.ID.eq(AUTHOR_BOOK.BOOK_ID))
            .innerJoin(AUTHOR).on(AUTHOR_BOOK.AUTHOR_ID.eq(AUTHOR.ID))
            .where(BOOK.ID.eq(id.value))
            .fetch()

        if (records.isEmpty()) {
            return null
        }

        val bookRecord = records.first()
        val authors = convertAuthorDtoList(records)

        return convert(bookRecord, authors)
    }

    /**
     * 著者情報を含む書籍レコードのリストから、著者DTOのリストを変換する
     */
    private fun convertAuthorDtoList(records: List<Record>): List<AuthorDto> {
        return records.map { record ->
            val authorId = record.get("author_id", Int::class.java)!!
            val authorName = record.get("author_name", String::class.java)!!
            val authorBirthDate = record.get("author_birth_date", java.time.LocalDate::class.java)
            val authorVersion = record.get("author_version", Int::class.java) ?: 1

            AuthorDto(
                id = ID(authorId),
                name = authorName,
                birthDate = authorBirthDate?.let { AuthorBirthDate(it) },
                version = authorVersion,
            )
        }
    }

    /**
     * 書籍レコードと著者DTOのリストから、書籍DTOを変換する
     */
    private fun convert(bookRecord: Record, authors: List<AuthorDto>): BookDto {
        return BookDto(
            id = ID(bookRecord[BOOK.ID]!!),
            title = bookRecord[BOOK.TITLE]!!,
            price = BookPrice.of(bookRecord[BOOK.PRICE]!!),
            authors = authors,
            status = BookPublishStatus.of(bookRecord[BOOK.PUBLISH_STATUS]!!),
            version = bookRecord[BOOK.VERSION]!!
        )
    }
}
