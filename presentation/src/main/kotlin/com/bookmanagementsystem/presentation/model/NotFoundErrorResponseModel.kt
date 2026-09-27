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
 * 汎用的なエラーレスポンス
 * @param errors kotlin.collections.List<ErrorModel>
 */
data class NotFoundErrorResponseModel(
    @field:Valid
    @get:JsonProperty("errors") val errors: kotlin.collections.List<ErrorModel>?
) {

}
