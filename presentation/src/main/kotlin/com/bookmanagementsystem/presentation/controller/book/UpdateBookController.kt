package com.bookmanagementsystem.presentation.controller.book

import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.presentation.model.UpdateBookRequestModel
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.register.UpdateBookCommand
import com.bookmanagementsystem.usecase.book.register.UpdateBookUsecase
import jakarta.validation.Valid
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RestController

@RestController
class UpdateBookController(private val usecase: UpdateBookUsecase) {
    /**
     * 書籍を更新する
     */
    @PutMapping("/api/books/{id}")
    fun update(@PathVariable id: Int, @Valid @RequestBody request: UpdateBookRequestModel): BookResponseModel {
        val dto = usecase.execute(toCommand(id, request))
        return toResponse(dto)
    }

    private fun toCommand(id: Int, request: UpdateBookRequestModel) = UpdateBookCommand(
        id = ID(id),
        title = request.title!!,
        price = BookPrice.of(request.price!!),
        authorIds = request.authorIds!!.map { ID(it) },
        status = BookPublishStatus.of(request.status!!.value)
    )

    private fun toResponse(dto: BookDto) = BookResponseModel(
        id = dto.id.value,
        title = dto.title,
        price = dto.price.toLong(),
        authors = dto.authors.map {
            AuthorResponseModel(
                it.id.value,
                it.name,
                it.birthDate?.value,
                it.version
            )
        },
        status = BookStatus.of(dto.status.value),
    )
}
