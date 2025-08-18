package com.bookmanagementsystem.infrastructure.book

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.BOOK_AUTHOR
import org.jooq.DSLContext

/**
 * BookAuthor関連のデータベース操作をまとめるヘルパーオブジェクト
 */
object BookAuthorHelper {
    /**
     * 書籍と著者の関連を挿入する
     */
    fun insertBookAuthor(dsl: DSLContext, bookId: ID<Book>, authorIds: List<ID<Author>>): List<ID<Author>> = dsl
        .insertInto(
            BOOK_AUTHOR,
            BOOK_AUTHOR.BOOK_ID,
            BOOK_AUTHOR.AUTHOR_ID
        )
        .apply { authorIds.forEach { values(bookId.value, it.value) } }
        .returning()
        .fetch()
        .map { ID(it[BOOK_AUTHOR.AUTHOR_ID]!!) }

    /**
     * 指定された書籍IDに関連する書籍-著者関連をすべて削除する
     */
    fun deleteBookAuthorByBookId(dsl: DSLContext, bookId: ID<Book>): Int = dsl
        .deleteFrom(BOOK_AUTHOR)
        .where(BOOK_AUTHOR.BOOK_ID.eq(bookId.value))
        .execute()
}
