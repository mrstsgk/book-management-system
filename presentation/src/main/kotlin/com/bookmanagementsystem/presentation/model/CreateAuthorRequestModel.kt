package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.Valid

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
    @get:JsonProperty("birthDate") val birthDate: java.time.LocalDate?
) {

}
