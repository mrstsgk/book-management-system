package com.bookmanagementsystem.domain.author

interface AuthorRepository {
    /**
     * 著者を登録する
     */
    fun insert(author: Author): Author
}
