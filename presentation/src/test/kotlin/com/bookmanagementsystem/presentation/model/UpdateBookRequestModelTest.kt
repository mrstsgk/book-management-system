package com.bookmanagementsystem.presentation.model

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.collections.shouldHaveSize
import io.kotest.matchers.shouldBe
import jakarta.validation.Validation
import jakarta.validation.Validator

class UpdateBookRequestModelTest : FunSpec({
    val validator: Validator = Validation.buildDefaultValidatorFactory().validator

    test("有効なパラメータでUpdateBookRequestModelを作成できる") {
        val title = "人間失格"
        val price = 1500L
        val authorIds = createDefaultAuthorIds()
        val status = BookStatus.PUBLISHED

        val requestModel = createValidUpdateBookRequest(
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
        val requestModel = UpdateBookRequestModel(
            title = null,
            price = 1500L,
            authorIds = createDefaultAuthorIds(),
            status = BookStatus.PUBLISHED,
            version = 1
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("価格がnullの場合バリデーションエラー") {
        val requestModel = UpdateBookRequestModel(
            title = "人間失格",
            price = null,
            authorIds = createDefaultAuthorIds(),
            status = BookStatus.PUBLISHED,
            version = 1
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("価格が負の値の場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            price = -1L
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Min.message}"
    }

    test("著者IDリストがnullの場合バリデーションエラー") {
        val requestModel = UpdateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = null,
            status = BookStatus.PUBLISHED,
            version = 1
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("著者IDリストが空の場合バリデーションエラー") {
        val requestModel = UpdateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = emptyList(),
            status = BookStatus.PUBLISHED,
            version = 1
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Size.message}"
    }

    test("ステータスがnullの場合バリデーションエラー") {
        val requestModel = UpdateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authorIds = createDefaultAuthorIds(),
            status = null,
            version = 1
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("複数の著者IDを設定できる") {
        val authorIds = listOf(1, 2)

        val requestModel = createValidUpdateBookRequest(
            title = "文学選集",
            price = 3000L,
            authorIds = authorIds
        )

        requestModel.authorIds?.shouldHaveSize(2)
        requestModel.authorIds shouldBe authorIds
    }

    test("出版済みステータスを設定できる") {
        val requestModel = createValidUpdateBookRequest(
            status = BookStatus.UNPUBLISHED
        )

        requestModel.status shouldBe BookStatus.UNPUBLISHED
        requestModel.status?.value shouldBe 1
    }

    test("未出版ステータスを設定できる") {
        val requestModel = createValidUpdateBookRequest(
            status = BookStatus.PUBLISHED
        )

        requestModel.status shouldBe BookStatus.PUBLISHED
        requestModel.status?.value shouldBe 2
    }

    test("価格の境界値テスト - 最小値0") {
        val requestModel = createValidUpdateBookRequest(
            title = "無料本",
            price = 0L
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 0
    }

    test("書籍情報を更新用に設定できる") {
        val requestModel = createValidUpdateBookRequest(
            title = "更新された人間失格",
            price = 2000L,
            authorIds = listOf(1, 2, 3),
            status = BookStatus.UNPUBLISHED
        )

        requestModel.title shouldBe "更新された人間失格"
        requestModel.price shouldBe 2000L
        requestModel.authorIds shouldBe listOf(1, 2, 3)
        requestModel.status shouldBe BookStatus.UNPUBLISHED
    }

    test("タイトルに半角スペースが含まれる場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            title = "人間 失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルに全角スペースが含まれる場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            title = "人間　失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルがタブ文字を含む場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            title = "人間\t失格"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルが半角スペースのみの場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            title = " "
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }

    test("タイトルが全角スペースのみの場合バリデーションエラー") {
        val requestModel = createValidUpdateBookRequest(
            title = "　"
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Pattern.message}"
    }
})

// テストフィクスチャ
fun createValidUpdateBookRequest(
    title: String = "人間失格",
    price: Long = 1500L,
    authorIds: List<Int> = createDefaultAuthorIds(),
    status: BookStatus = BookStatus.PUBLISHED,
    version: Int = 1
) = UpdateBookRequestModel(
    title = title,
    price = price,
    authorIds = authorIds,
    status = status,
    version = version
)
