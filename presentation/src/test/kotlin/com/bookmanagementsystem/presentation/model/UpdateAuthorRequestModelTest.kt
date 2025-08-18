package com.bookmanagementsystem.presentation.model

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import jakarta.validation.Validation
import jakarta.validation.Validator
import java.time.LocalDate

class UpdateAuthorRequestModelTest : FunSpec({
    val validator: Validator = Validation.buildDefaultValidatorFactory().validator

    test("有効なパラメータでUpdateAuthorRequestModelを作成できる") {
        val name = "太宰治"
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = UpdateAuthorRequestModel(
            name = name,
            version = 1,
            birthDate = birthDate
        )

        requestModel.name shouldBe name
        requestModel.birthDate shouldBe birthDate
    }

    test("生年月日がnullでも設定される") {
        val name = "夏目漱石"

        val requestModel = UpdateAuthorRequestModel(
            name = name,
            version = 1,
            birthDate = null
        )

        requestModel.name shouldBe name
        requestModel.birthDate shouldBe null
    }

    test("生年月日が当日日付ならエラー") {
        val name = "当日太郎"
        val birthDate = LocalDate.now()

        val requestModel = UpdateAuthorRequestModel(
            name = name,
            version = 1,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "生年月日は現在の日付よりも過去である必要があります"
    }

    test("名前がnullならエラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = UpdateAuthorRequestModel(
            name = null,
            version = 1,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        // デフォルトのロケールに依存しないようにメッセージテンプレートで検証
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("バージョンがnullならエラー") {
        val name = "太宰治"
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = UpdateAuthorRequestModel(
            name = name,
            version = null,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }
})
