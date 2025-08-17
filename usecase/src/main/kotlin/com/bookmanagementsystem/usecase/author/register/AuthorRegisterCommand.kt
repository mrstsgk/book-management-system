package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.core.ID

sealed interface AuthorRegisterCommand

/**
 * 著者を登録するコマンド
 */
data class CreateAuthorCommand(
    val name: String,
    val birthDate: AuthorBirthDate?
) : AuthorRegisterCommand

/**
 * 著者を更新するコマンド
 */
data class UpdateAuthorCommand(
    val id: ID<Author>,
    val name: String,
    val birthDate: AuthorBirthDate?
) : AuthorRegisterCommand
