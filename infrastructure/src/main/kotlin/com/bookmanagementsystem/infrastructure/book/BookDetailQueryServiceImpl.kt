package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.book.BookRecordConverter.convert
import com.bookmanagementsystem.infrastructure.book.BookRecordConverter.convertAuthorDtoList
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK_AUTHOR
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
            AUTHOR.VERSION.`as`("author_version")
        )
            .from(BOOK)
            // NOTE: 書籍には必ず著者が存在するというビジネスルールが強制されるため innerJoin で実装する
            .innerJoin(BOOK_AUTHOR).on(BOOK.ID.eq(BOOK_AUTHOR.BOOK_ID))
            .innerJoin(AUTHOR).on(BOOK_AUTHOR.AUTHOR_ID.eq(AUTHOR.ID))
            .where(BOOK.ID.eq(id.value))
            .fetch()

        // NOTE: 書籍自体が存在しない場合は null を返す
        if (records.isEmpty()) return null

        val bookRecord = records.first()
        val authors = convertAuthorDtoList(records)

        return convert(bookRecord, authors)
    }
}
