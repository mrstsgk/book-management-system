package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.constraints.Size
import jakarta.validation.constraints.Pattern
import jakarta.validation.Valid
import com.bookmanagementsystem.usecase.author.register.PastOnly

/**
 * 書籍のAPIで汎用的に使用するレスポンス情報
 * @param id 書籍ID
 * @param title 書籍タイトル
 * @param price 書籍価格
 * @param authors 著者のリスト
 * @param status BookStatus
 * @param version バージョン番号(楽観的ロック用)
 */
data class BookResponseModel(
    /* 書籍ID */
    @field:NotNull
    @get:JsonProperty("id") val id: kotlin.Int?,

    /* 書籍タイトル */
    @get:Pattern(regexp = "^[^\\s　]+$")
    @get:Size(max = 255)
    @field:NotNull
    @get:JsonProperty("title") val title: kotlin.String?,

    /* 書籍価格 */
    @get:Min(0L)
    @get:Max(99999999L)
    @field:NotNull
    @get:JsonProperty("price") val price: kotlin.Long?,

    /* 著者のリスト */
    @get:Size(min = 1)
    @field:Valid
    @field:NotNull
    @get:JsonProperty("authors") val authors: kotlin.collections.List<AuthorResponseModel>?,

    @field:Valid
    @field:NotNull
    @get:JsonProperty("status") val status: BookStatus?,

    /* バージョン番号(楽観的ロック用) */
    @field:NotNull
    @get:JsonProperty("version") val version: kotlin.Int?
) {

}
