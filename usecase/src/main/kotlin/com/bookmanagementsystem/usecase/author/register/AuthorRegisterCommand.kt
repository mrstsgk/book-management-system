package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.AuthorBirthDate

sealed interface AuthorRegisterCommand

/**
 * 著者を登録するコマンド
 */
data class CreateAuthorCommand(
    val name: String,
    val birthDate: AuthorBirthDate?
) : AuthorRegisterCommand
