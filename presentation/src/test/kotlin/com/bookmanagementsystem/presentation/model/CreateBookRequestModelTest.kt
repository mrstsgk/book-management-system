package com.bookmanagementsystem.presentation.model

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.collections.shouldHaveSize
import io.kotest.matchers.shouldBe
import jakarta.validation.Validation
import jakarta.validation.Validator

class CreateBookRequestModelTest : FunSpec({
    val validator: Validator = Validation.buildDefaultValidatorFactory().validator

    test("有効なパラメータでCreateBookRequestModelを作成できる") {
        val title = "人間失格"
        val price = 1500L
        val authorIds = createDefaultAuthorIds()
        val status = BookStatus.PUBLISHED

        val requestModel = createValidBookRequest(
            title = title,
            price = price,
            authorIds = authorIds,
            status = status
        )

        requestModel.title shouldBe title
        requestModel.price shouldBe price
        requestModel.authorIds shouldBe authorIds
        requestModel.status shouldBe status
    }

    test("タイトルがnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = null,
            price = 1500L,
            authorIds = createDefaultAuthorIds(),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("価格がnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = null,
            authorIds = createDefaultAuthorIds(),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("価格が負の値の場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            price = -1L
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Min.message}"
    }

    test("著者IDリストがnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = null,
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("著者IDリストが空の場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = emptyList(),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Size.message}"
    }

    test("ステータスがnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = createDefaultAuthorIds(),
            status = null
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("複数の著者IDを設定できる") {
        val authorIds = listOf(1, 2)

        val requestModel = createValidBookRequest(
            title = "文学選集",
            price = 3000L,
            authorIds = authorIds
        )

        requestModel.authorIds?.shouldHaveSize(2)
        requestModel.authorIds shouldBe authorIds
    }

    test("未出版ステータスを設定できる") {
        val requestModel = createValidBookRequest(
            status = BookStatus.UNPUBLISHED
        )

        requestModel.status shouldBe BookStatus.UNPUBLISHED
        requestModel.status?.value shouldBe 1
    }

    test("出版済みステータスを設定できる") {
        val requestModel = createValidBookRequest(
            status = BookStatus.PUBLISHED
        )

        requestModel.status shouldBe BookStatus.PUBLISHED
        requestModel.status?.value shouldBe 2
    }

    test("価格の境界値テスト - 最小値0") {
        val requestModel = createValidBookRequest(
            title = "無料本",
            price = 0L
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 0
    }

    test("タイトルが空文字列ならバリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = ""
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        // 空文字列は@Patternでエラーになる（@Size(max=255)のみで、minLengthは設定されていないため）
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルが最大長を超えるならバリデーションエラー") {
        val longTitle = "a".repeat(256) // 256文字
        val requestModel = createValidBookRequest(
            title = longTitle
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Size.message}"
    }

    test("タイトルが境界値（255文字）で有効") {
        val maxLengthTitle = "a".repeat(255) // 255文字
        val requestModel = createValidBookRequest(
            title = maxLengthTitle
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 0
    }

    test("タイトルに半角スペースが含まれる場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = "人間 失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルに全角スペースが含まれる場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = "人間　失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルがタブ文字を含む場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = "人間\t失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルが半角スペースのみの場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = " "
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルが全角スペースのみの場合バリデーションエラー") {
        val requestModel = createValidBookRequest(
            title = "　"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }
})

// テストフィクスチャ
fun createDefaultAuthorIds() = listOf(1)

fun createValidBookRequest(
    title: String = "人間失格",
    price: Long = 1500L,
    authorIds: List<Int> = createDefaultAuthorIds(),
    status: BookStatus = BookStatus.PUBLISHED
) = CreateBookRequestModel(
    title = title,
    price = price,
    authorIds = authorIds,
    status = status
)
