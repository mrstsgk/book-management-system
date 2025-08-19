package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK_AUTHOR
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookQueryService
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class BookQueryServiceImpl(private val dsl: DSLContext) : BookQueryService {
    override fun findByAuthorId(authorId: ID<Author>): List<BookDto> {
        val records = dsl.select(
            BOOK.ID,
            BOOK.TITLE,
            BOOK.PRICE,
            BOOK.PUBLISH_STATUS,
            BOOK.VERSION,
            AUTHOR.ID.`as`("author_id"),
            AUTHOR.NAME.`as`("author_name"),
            AUTHOR.BIRTH_DATE.`as`("author_birth_date"),
            AUTHOR.VERSION.`as`("author_version")
        )
            .from(BOOK)
            .innerJoin(BOOK_AUTHOR).on(BOOK.ID.eq(BOOK_AUTHOR.BOOK_ID))
            .innerJoin(AUTHOR).on(BOOK_AUTHOR.AUTHOR_ID.eq(AUTHOR.ID))
            .where(
                BOOK.ID.`in`(
                    dsl.select(BOOK_AUTHOR.BOOK_ID)
                        .from(BOOK_AUTHOR)
                        .where(BOOK_AUTHOR.AUTHOR_ID.eq(authorId.value))
                )
            )
            .orderBy(BOOK.ID)
            .fetch()

        // NOTE: 著者に紐づく書籍が存在しない場合は空のリストを返す
        if (records.isEmpty()) return emptyList()

        val bookGroups = records.groupBy { it[BOOK.ID] }
        return bookGroups.map { (_, bookRecords) ->
            val bookRecord = bookRecords.first()
            val authors = BookRecordConverter.convertAuthorDtoList(bookRecords)
            BookRecordConverter.convert(bookRecord, authors)
        }
    }
}
