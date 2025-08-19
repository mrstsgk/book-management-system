package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.read.BookDetailQueryService
import org.jooq.DSLContext
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
        val authors = BookRecordConverter.convertAuthorDtoList(records)

        return BookRecordConverter.convert(bookRecord, authors)
    }
}
