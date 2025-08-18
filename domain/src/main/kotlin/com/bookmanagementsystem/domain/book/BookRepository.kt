package com.bookmanagementsystem.domain.book

/**
 * 書籍のリポジトリインターフェース
 */
interface BookRepository {
    /**
     * 書籍を登録する
     */
    fun insert(book: Book): Book

    /**
     * 書籍を更新する
     */
    fun update(book: Book): Book
}
