package com.bookmanagementsystem.domain.author

import com.bookmanagementsystem.domain.core.ID

interface AuthorRepository {
    /**
     * IDに合う著者を取得する
     */
    fun findById(id: ID<Author>): Author?

    /**
     * 著者のIDリストに合う著者を取得する
     */
    fun findByIds(ids: List<ID<Author>>): List<Author>

    /**
     * 著者を登録する
     */
    fun insert(author: Author): Author

    /**
     * 著者を更新する
     */
    fun update(author: Author): Author
}
