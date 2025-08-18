package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.records.BookRecord
import com.bookmanagementsystem.jooq.tables.references.BOOK
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

        val authorIds = BookAuthorHelper.insertBookAuthor(dsl, ID(record.id!!), book.authorIds)
        return convert(record, authorIds)
    }

    override fun update(book: Book): Book {
        val record = dsl.update(BOOK)
            .set(BOOK.TITLE, book.title)
            .set(BOOK.PRICE, book.price.amount.value)
            .set(BOOK.PUBLISH_STATUS, book.status.value)
            .where(BOOK.ID.eq(book.id!!.value))
            .returning()
            .fetchSingle()

        // 既存の著者関連を削除して再挿入
        BookAuthorHelper.deleteBookAuthorByBookId(dsl, book.id!!)

        val authorIds = BookAuthorHelper.insertBookAuthor(dsl, book.id!!, book.authorIds)
        return convert(record, authorIds)
    }

    private fun convert(record: BookRecord, authorIds: List<ID<Author>>) = Book(
        id = ID(record.id!!),
        title = record.title!!,
        price = BookPrice.of(record.price!!),
        authorIds = authorIds,
        status = BookPublishStatus.of(record.publishStatus!!),
    )
}
