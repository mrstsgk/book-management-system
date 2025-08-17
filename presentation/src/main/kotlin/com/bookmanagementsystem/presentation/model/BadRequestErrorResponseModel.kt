package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.Valid
import com.bookmanagementsystem.usecase.validation.PastOnly

/**
 * リクエストのバリデーションエラー
 */
data class BadRequestErrorResponseModel(
    @field:Valid
    @get:JsonProperty("errors") val errors: kotlin.collections.List<ErrorModel>?
) {

}
