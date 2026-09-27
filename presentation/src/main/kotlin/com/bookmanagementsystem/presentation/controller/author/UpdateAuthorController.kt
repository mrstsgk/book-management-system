package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.UpdateAuthorRequestModel
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.register.UpdateAuthorCommand
import com.bookmanagementsystem.usecase.author.register.UpdateAuthorUsecase
import jakarta.validation.Valid
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RestController

@RestController
class UpdateAuthorController(private val usecase: UpdateAuthorUsecase) {
    /**
     * 著者を更新する
     */
    @PutMapping("/api/authors/{id}")
    fun update(@PathVariable id: Int, @Valid @RequestBody request: UpdateAuthorRequestModel): AuthorResponseModel {
        val dto = usecase.execute(toCommand(id, request))
        return toResponse(dto)
    }

    private fun toCommand(id: Int, request: UpdateAuthorRequestModel) = UpdateAuthorCommand(
        id = ID(id),
        name = request.name!!,
        birthDate = request.birthDate?.let { AuthorBirthDate(it) },
        version = request.version!!,
    )

    private fun toResponse(dto: AuthorDto) = AuthorResponseModel(
        id = dto.id.value,
        name = dto.name,
        birthDate = dto.birthDate?.value,
        version = dto.version
    )
}
