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
 * 著者更新APIのリクエスト情報
 * @param name 著者名
 * @param version バージョン番号(楽観的ロック用)
 * @param birthDate 生年月日
 */
data class UpdateAuthorRequestModel(
    /* 著者名 */
    @get:Pattern(regexp = "^[^\\s　]+$")
    @get:Size(max = 100)
    @field:NotNull
    @get:JsonProperty("name") val name: kotlin.String?,

    /* バージョン番号(楽観的ロック用) */
    @field:NotNull
    @get:JsonProperty("version") val version: kotlin.Int?,

    /* 生年月日 */
    @field:Valid
    @field:PastOnly
    @get:JsonProperty("birthDate") val birthDate: java.time.LocalDate?
) {

}
