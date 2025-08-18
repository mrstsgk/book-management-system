package com.bookmanagementsystem.domain.book

import com.bookmanagementsystem.domain.core.ID

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

    /**
     * 書籍IDから該当書籍のステータスを取得する
     *
     * NOTE: なぜ get なのか？
     * 書籍の存在はこのユースケース内で保証されており、存在しないことはドメイン上ありえない前提であるため、
     * Optional や null を返す `find` ではなく、存在が確定している `get` を用いている。
     * 万が一存在しなければシステムの不整合として例外を投げることで早期に検知する意図がある。
     */
    fun getBookPublishStatusById(id: ID<Book>): BookPublishStatus
}
