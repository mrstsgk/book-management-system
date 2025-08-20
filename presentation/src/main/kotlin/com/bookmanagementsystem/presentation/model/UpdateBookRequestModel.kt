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
 * 書籍更新APIのリクエスト情報
 * @param title 書籍タイトル
 * @param price 書籍価格
 * @param authorIds 著者IDのリスト
 * @param status BookStatus
 * @param version バージョン番号(楽観的ロック用)
 */
data class UpdateBookRequestModel(
    /* 書籍タイトル */
    @get:Pattern(regexp = "^[^\\s　]+$")
    @get:Size(max = 255)
    @field:NotNull
    @get:JsonProperty("title") val title: kotlin.String?,

    /* 書籍価格 */
    @get:Min(0L)
    @get:Max(9999999999L)
    @field:NotNull
    @get:JsonProperty("price") val price: kotlin.Long?,

    /* 著者IDのリスト */
    @get:Size(min = 1)
    @field:NotNull
    @get:JsonProperty("authorIds") val authorIds: kotlin.collections.List<kotlin.Int>?,

    @field:Valid
    @field:NotNull
    @get:JsonProperty("status") val status: BookStatus?,

    /* バージョン番号(楽観的ロック用) */
    @field:NotNull
    @get:JsonProperty("version") val version: kotlin.Int?
) {

}
