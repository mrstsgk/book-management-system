package com.bookmanagementsystem.presentation.model

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import jakarta.validation.Validation
import jakarta.validation.Validator
import java.time.LocalDate

class CreateAuthorRequestModelTest : FunSpec({
    val validator: Validator = Validation.buildDefaultValidatorFactory().validator

    test("有効なパラメータでCreateAuthorRequestModelを作成できる") {
        val name = "太宰治"
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = name,
            birthDate = birthDate
        )

        requestModel.name shouldBe name
        requestModel.birthDate shouldBe birthDate
    }

    test("生年月日がnullでも設定される") {
        val name = "夏目漱石"

        val requestModel = CreateAuthorRequestModel(
            name = name,
            birthDate = null
        )

        requestModel.name shouldBe "夏目漱石"
        requestModel.birthDate shouldBe null
    }

    test("名前がnullならエラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = null,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().message shouldBe "null は許可されていません"
    }
})
