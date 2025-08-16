package com.bookmanagementsystem.domain.core

/**
 * エンティティのIDを型安全に表現する値オブジェクト.
 *
 * ID<Book> や ID<Author> のように、特定のエンティティに紐づくIDを表現する。
 */
data class ID<E>(val value: Int) {
    init {
        require(value > 0)
    }
}
