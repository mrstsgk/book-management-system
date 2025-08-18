package com.bookmanagementsystem.usecase.author.register

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.usecase.author.AuthorDto
import org.springframework.stereotype.Service

@Service
class UpdateAuthorUsecase(private val repository: AuthorRepository) {
    /**
     * 著者を更新する
     */
    fun execute(command: UpdateAuthorCommand): AuthorDto {
        repository.findById(command.id) ?: throw NoSuchElementException("著者が見つかりません: ${command.id}")

        val author = repository.update(toEntity(command))
        return toDto(author)
    }

    private fun toEntity(command: UpdateAuthorCommand) = Author(
        id = command.id,
        name = command.name,
        birthDate = command.birthDate,
        version = command.version
    )

    private fun toDto(author: Author) = AuthorDto(
        id = author.id!!,
        name = author.name,
        birthDate = author.birthDate,
        version = author.version!!
    )
}
