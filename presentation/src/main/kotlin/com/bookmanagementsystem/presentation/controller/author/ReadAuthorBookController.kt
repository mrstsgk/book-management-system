package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.usecase.author.read.ReadAuthorBookUsecase
import com.bookmanagementsystem.usecase.book.BookDto
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RestController

@RestController
class ReadAuthorBookController(private val usecase: ReadAuthorBookUsecase) {
    /**
     * 著者に紐づく書籍一覧を取得する
     */
    @GetMapping("/api/authors/{id}/books")
    fun read(@PathVariable id: Int): List<BookResponseModel> {
        val bookDtoList = usecase.execute(ID(id))

        return bookDtoList.map { toResponse(it) }
    }

    private fun toResponse(dto: BookDto) = BookResponseModel(
        id = dto.id.value,
        title = dto.title,
        price = dto.price.toLong(),
        authors = dto.authors.map {
            AuthorResponseModel(
                id = it.id.value,
                name = it.name,
                birthDate = it.birthDate?.value
            )
        },
        status = BookStatus.of(dto.status.value)
    )
}
