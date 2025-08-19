package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK
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
            .innerJoin(AUTHOR_BOOK).on(BOOK.ID.eq(AUTHOR_BOOK.BOOK_ID))
            .innerJoin(AUTHOR).on(AUTHOR_BOOK.AUTHOR_ID.eq(AUTHOR.ID))
            .where(
                BOOK.ID.`in`(
                    dsl.select(AUTHOR_BOOK.BOOK_ID)
                        .from(AUTHOR_BOOK)
                        .where(AUTHOR_BOOK.AUTHOR_ID.eq(authorId.value))
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
