package com.bookmanagementsystem.usecase.validation.core

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.core.ID
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe

/**
 * UniqueIdValidatorのテスト
 */
class UniqueIdValidatorTest : FunSpec({
    val validator = UniqueIdValidator()

    context("ID重複チェックのバリデーション") {
        test("重複のないIDリストの場合はtrue") {
            val uniqueIds: List<ID<Author>> = listOf(ID(1), ID(2), ID(3))
            val result = validator.isValid(uniqueIds, null)

            result shouldBe true
        }

        test("重複のあるIDリストの場合はfalse") {
            val duplicateIds: List<ID<Author>> = listOf(ID(1), ID(2), ID(1))
            val result = validator.isValid(duplicateIds, null)

            result shouldBe false
        }

        test("空のIDリストの場合はtrue") {
            val emptyIds = emptyList<ID<Author>>()
            val result = validator.isValid(emptyIds, null)

            result shouldBe true
        }

        test("単一のIDの場合はtrue") {
            val singleId = listOf(ID<Author>(1))
            val result = validator.isValid(singleId, null)

            result shouldBe true
        }

        test("同じIDが複数回含まれる場合はfalse") {
            val sameIds: List<ID<Author>> = listOf(ID(5), ID(5), ID(5))
            val result = validator.isValid(sameIds, null)

            result shouldBe false
        }

        test("連続する重複IDがある場合はfalse") {
            val consecutiveDuplicates: List<ID<Author>> = listOf(ID(1), ID(1), ID(2))
            val result = validator.isValid(consecutiveDuplicates, null)

            result shouldBe false
        }

        test("非連続の重複IDがある場合はfalse") {
            val nonConsecutiveDuplicates: List<ID<Author>> = listOf(ID(1), ID(2), ID(3), ID(1))
            val result = validator.isValid(nonConsecutiveDuplicates, null)

            result shouldBe false
        }
    }
})
