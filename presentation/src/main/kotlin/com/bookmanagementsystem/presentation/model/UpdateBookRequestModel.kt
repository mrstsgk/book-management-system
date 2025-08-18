package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.Valid
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Size

/**
 * 書籍更新APIのリクエスト情報
 * @param title 書籍タイトル
 * @param price 書籍価格
 * @param authorIds 著者IDのリスト
 * @param status BookStatus
 */
data class UpdateBookRequestModel(
    /* 書籍タイトル */
    @field:NotNull
    @get:JsonProperty("title") val title: kotlin.String?,

    /* 書籍価格 */
    @get:Min(0L)
    @field:NotNull
    @get:JsonProperty("price") val price: kotlin.Long?,

    /* 著者IDのリスト */
    @get:Size(min = 1)
    @field:NotNull
    @get:JsonProperty("authorIds") val authorIds: kotlin.collections.List<kotlin.Int>?,

    @field:Valid
    @field:NotNull
    @get:JsonProperty("status") val status: BookStatus?
) {

}
