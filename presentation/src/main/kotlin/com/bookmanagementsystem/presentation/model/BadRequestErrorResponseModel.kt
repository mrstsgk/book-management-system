package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.Valid

/**
 * リクエストのバリデーションエラー
 * @param errors kotlin.collections.List<ErrorModel>
 */
data class BadRequestErrorResponseModel(
    @field:Valid
    @get:JsonProperty("errors") val errors: kotlin.collections.List<ErrorModel>?
) {

}
