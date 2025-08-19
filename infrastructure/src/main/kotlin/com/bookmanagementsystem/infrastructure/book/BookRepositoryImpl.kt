package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.exception.OptimisticLockException
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
            .set(BOOK.VERSION, 1)
            .returning()
            .fetchSingle()

        val authorIds = AuthorBookHelper.insertAuthorBook(dsl, ID(record.id!!), book.authorIds)
        return convert(record, authorIds)
    }

    override fun update(book: Book): Book {
        val record = dsl
            .update(BOOK)
            .set(BOOK.TITLE, book.title)
            .set(BOOK.PRICE, book.price.amount.value)
            .set(BOOK.PUBLISH_STATUS, book.status.value)
            .set(BOOK.VERSION, book.version!! + 1)
            .where(BOOK.ID.eq(book.id!!.value))
            .and(BOOK.VERSION.eq(book.version!!))
            .returning()
            .fetchOne()
            ?: throw OptimisticLockException("書籍の更新に失敗しました。")

        // 既存の著者関連を削除して再挿入
        AuthorBookHelper.deleteAuthorBookByBookId(dsl, book.id!!)

        val authorIds = AuthorBookHelper.insertAuthorBook(dsl, book.id!!, book.authorIds)
        return convert(record, authorIds)
    }

    override fun getBookPublishStatusById(id: ID<Book>): BookPublishStatus =
        BookPublishStatus.of(
            dsl.select(BOOK.PUBLISH_STATUS)
                .from(BOOK)
                .where(BOOK.ID.eq(id.value))
                .fetchSingleInto(Int::class.java)
        )

    private fun convert(record: BookRecord, authorIds: List<ID<Author>>) = Book(
        id = ID(record.id!!),
        title = record.title!!,
        price = BookPrice.of(record.price!!),
        authorIds = authorIds,
        status = BookPublishStatus.of(record.publishStatus!!),
        version = record.version
    )
}
