package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import org.jooq.DSLContext

/**
 * BookAuthor関連のデータベース操作をまとめるヘルパーオブジェクト
 */
object AuthorBookHelper {
    /**
     * 書籍と著者の関連を挿入する
     */
    fun insertAuthorBook(dsl: DSLContext, bookId: ID<Book>, authorIds: List<ID<Author>>): List<ID<Author>> = dsl
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
    fun deleteAuthorBookByBookId(dsl: DSLContext, bookId: ID<Book>): Int = dsl
        .deleteFrom(AUTHOR_BOOK)
        .where(AUTHOR_BOOK.BOOK_ID.eq(bookId.value))
        .execute()
}
