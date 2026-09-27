package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.usecase.book.BookSummaryDto
import com.bookmanagementsystem.usecase.book.read.BookQueryService
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class BookQueryServiceImpl(private val dsl: DSLContext) : BookQueryService {
    override fun findByAuthorId(authorId: ID<Author>): List<BookSummaryDto> {
        val records = dsl.selectDistinct(
            BOOK.ID,
            BOOK.TITLE,
            BOOK.PRICE,
            BOOK.PUBLISH_STATUS,
        )
            .from(BOOK)
            .innerJoin(AUTHOR_BOOK).on(BOOK.ID.eq(AUTHOR_BOOK.BOOK_ID))
            .where(AUTHOR_BOOK.AUTHOR_ID.eq(authorId.value))
            .orderBy(BOOK.ID)
            .fetch()

        return records.map { record ->
            BookSummaryDto(
                id = ID(record[BOOK.ID]!!),
                title = record[BOOK.TITLE]!!,
                price = BookPrice.of(record[BOOK.PRICE]!!),
                status = BookPublishStatus.of(record[BOOK.PUBLISH_STATUS]!!),
            )
        }
    }
}
