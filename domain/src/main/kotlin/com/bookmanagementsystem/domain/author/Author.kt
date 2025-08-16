package com.bookmanagementsystem.domain.author

import com.bookmanagementsystem.domain.core.ID

/**
 * 著者を表現するエンティティ
 */
data class Author(
    val id: ID<Author>,
    val name: String,
    val birthDate: AuthorBirthDate,
)
