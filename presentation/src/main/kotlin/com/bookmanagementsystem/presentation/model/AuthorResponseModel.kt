package com.bookmanagementsystem.presentation.model

import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Min
import jakarta.validation.constraints.Max
import jakarta.validation.Valid
import com.bookmanagementsystem.usecase.validation.PastOnly

/**
 * 著者のAPIで汎用的に使用するレスポンス情報
 * @param id 著者ID
 * @param name 著者名
 * @param birthDate 生年月日
 */
data class AuthorResponseModel(

    /* 著者ID */
    @field:NotNull
    @get:JsonProperty("id") val id: kotlin.Int?,

    /* 著者名 */
    @field:NotNull
    @get:JsonProperty("name") val name: kotlin.String?,

    /* 生年月日 */
    @field:Valid
    @get:JsonProperty("birthDate") val birthDate: java.time.LocalDate?
) {

}
