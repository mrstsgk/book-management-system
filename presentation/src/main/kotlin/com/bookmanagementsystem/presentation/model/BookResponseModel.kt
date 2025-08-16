package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.Valid

/**
 * 書籍のAPIで汎用的に使用するレスポンス情報
 * @param id 書籍ID
 * @param title 書籍タイトル
 * @param price 価格（円）(10桁)
 */
data class BookResponseModel(

    /* 書籍ID */
    @field:NotNull
    @get:JsonProperty("id") val id: kotlin.Int?,

    /* 書籍タイトル */
    @field:NotNull
    @get:JsonProperty("title") val title: kotlin.String?,

    /* 価格（円）(10桁) */
    @get:Min(0L)
    @get:Max(9999999999L)
    @field:NotNull
    @get:JsonProperty("price") val price: kotlin.Long?,
    @field:Valid
    @field:NotNull
    @get:JsonProperty("status") val status: BookStatus?
) {

}
