package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty

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
