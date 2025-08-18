package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.usecase.author.AuthorDto
import org.springframework.stereotype.Service

@Service
class CreateAuthorUsecase(private val repository: AuthorRepository) {
    /**
     * 著者を登録する
     */
    fun execute(command: CreateAuthorCommand): AuthorDto {
        val author = repository.insert(toEntity(command))
        return toDto(author)
    }

    private fun toEntity(command: CreateAuthorCommand) = Author(
        name = command.name,
        birthDate = command.birthDate
    )

    private fun toDto(author: Author) = AuthorDto(
        id = author.id!!,
        name = author.name,
        birthDate = author.birthDate,
        version = author.version!!
    )
}
