package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.book.BookRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.exception.OptimisticLockException
import com.bookmanagementsystem.jooq.tables.records.BookRecord
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.jooq.tables.references.BOOK
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class BookRepositoryImpl(private val dsl: DSLContext) : BookRepository {
    override fun findById(id: ID<Book>): Book? {
        val record = dsl.selectFrom(BOOK)
            .where(BOOK.ID.eq(id.value))
            .fetchOne() ?: return null

        val authorIds = findAuthorIdsByBookId(ID(record.id!!))
        // NOTE: 書籍に紐づく著者が存在しない場合、データ不整合が発生しているため例外をスロー
        check(authorIds.isNotEmpty())

        return convert(record, authorIds)
    }

    override fun insert(book: Book): Book {
        val record = dsl.insertInto(BOOK)
            .set(BOOK.TITLE, book.title)
            .set(BOOK.PRICE, book.price.amount.value)
            .set(BOOK.PUBLISH_STATUS, book.status.value)
            .set(BOOK.VERSION, 1)
            .returning()
            .fetchSingle()

        val authorIds = insertAuthorBook(ID(record.id!!), book.authorIds)
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
        deleteAuthorBookByBookId(book.id!!)

        val authorIds = insertAuthorBook(book.id!!, book.authorIds)
        return convert(record, authorIds)
    }

    override fun getBookPublishStatusById(id: ID<Book>): BookPublishStatus =
        BookPublishStatus.of(
            dsl.select(BOOK.PUBLISH_STATUS)
                .from(BOOK)
                .where(BOOK.ID.eq(id.value))
                .fetchSingleInto(Int::class.java)
        )

    /**
     * 指定された書籍IDに関連する著者IDのリストを取得する
     */
    private fun findAuthorIdsByBookId(bookId: ID<Book>): List<ID<Author>> = dsl
        .select(AUTHOR_BOOK.AUTHOR_ID)
        .from(AUTHOR_BOOK)
        .where(AUTHOR_BOOK.BOOK_ID.eq(bookId.value))
        .fetch()
        .map { ID(it[AUTHOR_BOOK.AUTHOR_ID]!!) }

    /**
     * 書籍と著者の関連を挿入する
     */
    private fun insertAuthorBook(bookId: ID<Book>, authorIds: List<ID<Author>>): List<ID<Author>> = dsl
        .insertInto(
            AUTHOR_BOOK,
            AUTHOR_BOOK.BOOK_ID,
            AUTHOR_BOOK.AUTHOR_ID,
            AUTHOR_BOOK.VERSION
        )
        .apply { authorIds.forEach { values(bookId.value, it.value, 1) } }
        .returning()
        .fetch()
        .map { ID(it[AUTHOR_BOOK.AUTHOR_ID]!!) }

    /**
     * 指定された書籍IDに関連する書籍-著者関連をすべて削除する
     */
    private fun deleteAuthorBookByBookId(bookId: ID<Book>): Int = dsl
        .deleteFrom(AUTHOR_BOOK)
        .where(AUTHOR_BOOK.BOOK_ID.eq(bookId.value))
        .execute()

    private fun convert(record: BookRecord, authorIds: List<ID<Author>>) = Book(
        id = ID(record.id!!),
        title = record.title!!,
        price = BookPrice.of(record.price!!),
        authorIds = authorIds,
        status = BookPublishStatus.of(record.publishStatus!!),
        version = record.version
    )
}
