package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.CreateAuthorRequestModel
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.register.CreateAuthorCommand
import com.bookmanagementsystem.usecase.author.register.CreateAuthorUsecase
import jakarta.validation.Valid
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RestController

@RestController
class CreateAuthorController(private val usecase: CreateAuthorUsecase) {
    /**
     * 著者を登録する
     */
    @PostMapping("/api/authors")
    fun create(@Valid @RequestBody request: CreateAuthorRequestModel): AuthorResponseModel {
        val dto = usecase.execute(toCommand(request))
        return toResponse(dto)
    }

    private fun toCommand(request: CreateAuthorRequestModel) = CreateAuthorCommand(
        name = request.name!!,
        birthDate = request.birthDate?.let { AuthorBirthDate(it) }
    )

    private fun toResponse(dto: AuthorDto): AuthorResponseModel {
        return AuthorResponseModel(
            id = dto.id.value,
            name = dto.name,
            birthDate = dto.birthDate?.value,
        )
    }
}
