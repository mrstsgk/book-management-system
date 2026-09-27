package com.bookmanagementsystem.usecase.author

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.core.ID

/**
 * 著者Dto
 */
data class AuthorDto(
    val id: ID<Author>,
    val name: String,
    val birthDate: AuthorBirthDate?,
    val version: Int,
)
