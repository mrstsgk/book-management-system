package com.bookmanagementsystem.presentation.controller.author.book

import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.presentation.model.AuthorBookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.usecase.author.book.read.ReadAuthorBookUsecase
import com.bookmanagementsystem.usecase.book.BookSummaryDto
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RestController

@RestController
class ReadAuthorBookController(private val usecase: ReadAuthorBookUsecase) {
    /**
     * 著者に紐づく書籍一覧を取得する
     *
     * NOTE: 書籍情報で著者情報とバージョンを返さない理由
     * - 著者ID指定での書籍取得時はauthorsフィールドは冗長のため除外
     * - バージョン情報は更新操作時のみ必要であり、一覧取得時には不要なため
     *
     */
    @GetMapping("/api/authors/{id}/books")
    fun read(@PathVariable id: Int): List<AuthorBookResponseModel> {
        val bookDtoList = usecase.execute(ID(id))

        return bookDtoList.map { toResponse(it) }
    }

    private fun toResponse(dto: BookSummaryDto) = AuthorBookResponseModel(
        id = dto.id.value,
        title = dto.title,
        price = dto.price.toLong(),
        status = BookStatus.of(dto.status.value),
    )
}
