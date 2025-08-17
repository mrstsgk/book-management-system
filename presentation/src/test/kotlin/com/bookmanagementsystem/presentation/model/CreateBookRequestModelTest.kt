package com.bookmanagementsystem.presentation.model

import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.collections.shouldHaveSize
import io.kotest.matchers.shouldBe
import jakarta.validation.Validation
import jakarta.validation.Validator
import java.time.LocalDate

class CreateBookRequestModelTest : FunSpec({
    val validator: Validator = Validation.buildDefaultValidatorFactory().validator

    test("有効なパラメータでCreateBookRequestModelを作成できる") {
        val title = "人間失格"
        val price = 1500L
        val authors = listOf(
            AuthorResponseModel(
                id = 1,
                name = "太宰治",
                birthDate = LocalDate.of(1909, 6, 19)
            )
        )
        val status = BookStatus.PUBLISHED

        val requestModel = CreateBookRequestModel(
            title = title,
            price = price,
            authors = authors,
            status = status
        )

        requestModel.title shouldBe title
        requestModel.price shouldBe price
        requestModel.authors shouldBe authors
        requestModel.status shouldBe status
    }

    test("タイトルがnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = null,
            price = 1500L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
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
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("価格が負の値の場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = -1L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Min.message}"
    }

    test("価格が最大値を超える場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 10000000000L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.Max.message}"
    }

    test("著者リストがnullの場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authors = null,
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("著者リストが空の場合バリデーションエラー") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authors = emptyList(),
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
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = null
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 1
        violations.first().messageTemplate shouldBe "{jakarta.validation.constraints.NotNull.message}"
    }

    test("複数の著者を設定できる") {
        val authors = listOf(
            AuthorResponseModel(
                id = 1,
                name = "太宰治",
                birthDate = LocalDate.of(1909, 6, 19)
            ),
            AuthorResponseModel(
                id = 2,
                name = "夏目漱石",
                birthDate = LocalDate.of(1867, 2, 9)
            )
        )

        val requestModel = CreateBookRequestModel(
            title = "文学選集",
            price = 3000L,
            authors = authors,
            status = BookStatus.PUBLISHED
        )

        requestModel.authors?.shouldHaveSize(2)
        requestModel.authors shouldBe authors
    }

    test("出版済みステータスを設定できる") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        requestModel.status shouldBe BookStatus.PUBLISHED
        requestModel.status?.value shouldBe 1
    }

    test("未出版ステータスを設定できる") {
        val requestModel = CreateBookRequestModel(
            title = "人間失格",
            price = 1500L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.UNPUBLISHED
        )

        requestModel.status shouldBe BookStatus.UNPUBLISHED
        requestModel.status?.value shouldBe 2
    }

    test("価格の境界値テスト - 最小値0") {
        val requestModel = CreateBookRequestModel(
            title = "無料本",
            price = 0L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 0
    }

    test("価格の境界値テスト - 最大値9999999999") {
        val requestModel = CreateBookRequestModel(
            title = "高額本",
            price = 9999999999L,
            authors = listOf(
                AuthorResponseModel(
                    id = 1,
                    name = "太宰治",
                    birthDate = LocalDate.of(1909, 6, 19)
                )
            ),
            status = BookStatus.PUBLISHED
        )

        val violations = validator.validate(requestModel)
        violations shouldHaveSize 0
    }
})
