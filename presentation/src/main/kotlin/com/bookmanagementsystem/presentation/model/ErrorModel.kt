package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.constraints.Size
import jakarta.validation.Valid
import com.bookmanagementsystem.usecase.validation.PastOnly

/**
 * エラー情報
 * @param code エラーコード
 * @param message エラーメッセージ
 */
data class ErrorModel(
    /* エラーコード */
    @get:JsonProperty("code") val code: kotlin.String?,

    /* エラーメッセージ */
    @get:JsonProperty("message") val message: kotlin.String?
) {

}
