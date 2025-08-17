package com.bookmanagementsystem.presentation.controller.book

import com.bookmanagementsystem.domain.book.BookPrice
import com.bookmanagementsystem.domain.book.BookPublishStatus
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.presentation.model.CreateBookRequestModel
import com.bookmanagementsystem.usecase.book.BookDto
import com.bookmanagementsystem.usecase.book.register.CreateBookCommand
import com.bookmanagementsystem.usecase.book.register.CreateBookUsecase
import jakarta.validation.Valid
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RestController

@RestController
class CreateBookController(private val usecase: CreateBookUsecase) {
    /**
     * 書籍を登録する
     */
    @PostMapping("/api/books")
    fun create(@Valid @RequestBody request: CreateBookRequestModel): BookResponseModel {
        val dto = usecase.execute(toCommand(request))
        return toResponse(dto)
    }

    private fun toCommand(request: CreateBookRequestModel) = CreateBookCommand(
        title = request.title!!,
        price = BookPrice.of(request.price!!),
        authorIds = request.authorIds!!.map { ID(it) },
        status = BookPublishStatus.of(request.status!!.value),
    )

    private fun toResponse(dto: BookDto) = BookResponseModel(
        id = dto.id.value,
        title = dto.title,
        price = dto.price.toLong(),
        authors = dto.authors.map {
            AuthorResponseModel(
                it.id.value,
                it.name,
                it.birthDate?.value
            )
        },
        status = BookStatus.of(dto.status.value),
    )
}
