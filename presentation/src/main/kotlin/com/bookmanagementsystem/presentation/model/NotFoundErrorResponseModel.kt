package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.constraints.Past
import jakarta.validation.Valid
import com.bookmanagementsystem.usecase.validation.PastOnly

/**
 * 汎用的なエラーレスポンス
 */
data class NotFoundErrorResponseModel(
    @field:Valid
    @get:JsonProperty("errors") val errors: kotlin.collections.List<ErrorModel>?
) {

}
