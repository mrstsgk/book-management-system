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

        requestModel.name shouldBe name
        requestModel.birthDate shouldBe null
    }

    test("生年月日が当日日付ならエラー") {
        val name = "当日太郎"
        val birthDate = LocalDate.now()

        val requestModel = CreateAuthorRequestModel(
            name = name,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "生年月日は現在の日付よりも過去である必要があります"
    }

    test("名前がnullならエラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = null,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        // デフォルトのロケールに依存しないようにメッセージテンプレートで検証
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("名前が空文字列ならエラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = "",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        // 空文字列は@Patternでエラーになる（@Size(max=100)のみで、minLengthは設定されていないため）
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("名前が最大長を超えるならエラー") {
        val birthDate = LocalDate.of(1909, 6, 19)
        val longName = "a".repeat(101) // 101文字

        val requestModel = CreateAuthorRequestModel(
            name = longName,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Size.message}"
    }

    test("名前が境界値（100文字）で有効") {
        val birthDate = LocalDate.of(1909, 6, 19)
        val maxLengthName = "a".repeat(100) // 100文字

        val requestModel = CreateAuthorRequestModel(
            name = maxLengthName,
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 0
    }

    test("名前に半角スペースが含まれる場合エラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = "太宰 治",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("名前に全角スペースが含まれる場合エラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = "太宰　治",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("名前がタブ文字を含む場合エラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = "太宰\t治",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("名前が半角スペースのみの場合エラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = " ",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("名前が全角スペースのみの場合エラー") {
        val birthDate = LocalDate.of(1909, 6, 19)

        val requestModel = CreateAuthorRequestModel(
            name = "　",
            birthDate = birthDate
        )

        val result = validator.validate(requestModel)
        result.size shouldBe 1
        result.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }
})
