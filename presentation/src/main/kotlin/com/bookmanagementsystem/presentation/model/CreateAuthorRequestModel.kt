package com.bookmanagementsystem.presentation.model

import com.bookmanagementsystem.usecase.author.register.PastOnly
import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.Valid
import jakarta.validation.constraints.NotNull

/**
 * 著者登録APIのリクエスト情報
 * @param name 著者名
 * @param birthDate 生年月日
 */
data class CreateAuthorRequestModel(
    /* 著者名 */
    @field:NotNull
    @get:JsonProperty("name") val name: kotlin.String?,

    /* 生年月日 */
    @field:Valid
    @field:PastOnly
    @get:JsonProperty("birthDate") val birthDate: java.time.LocalDate?
) {

}
