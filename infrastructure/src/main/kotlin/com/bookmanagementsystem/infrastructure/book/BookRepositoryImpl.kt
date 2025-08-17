package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.records.BookRecord
import com.bookmanagementsystem.jooq.tables.references.BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK_AUTHOR
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class BookRepositoryImpl(private val dsl: DSLContext) : BookRepository {
    override fun insert(book: Book): Book {
        val record = dsl.insertInto(BOOK)
            .set(BOOK.TITLE, book.title)
            .set(BOOK.PRICE, book.price.amount.value)
            .set(BOOK.PUBLISH_STATUS, book.status.value)
            .returning()
            .fetchSingle()

        val authorIds = insertBookAuthor(ID(record.id!!), book.authorIds)
        return convert(record, authorIds)
    }

    private fun insertBookAuthor(bookId: ID<Book>, authorIds: List<ID<Author>>): List<ID<Author>> = dsl
        .insertInto(
            BOOK_AUTHOR,
            BOOK_AUTHOR.BOOK_ID,
            BOOK_AUTHOR.AUTHOR_ID
        )
        .apply { authorIds.forEach { values(bookId.value, it.value) } }
        .returning()
        .fetch()
        .map { ID(it[BOOK_AUTHOR.AUTHOR_ID]!!) }

    private fun convert(record: BookRecord, authorIds: List<ID<Author>>) = Book(
        id = ID(record.id!!),
        title = record.title!!,
        price = BookPrice.of(record.price!!),
        authorIds = authorIds,
        status = BookPublishStatus.of(record.publishStatus!!),
    )
}
