package com.bookmanagementsystem.domain.book

import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe

class BookPublishStatusTest : FunSpec({
    context("値からのBookPublishStatus生成") {
        test("値1からUNPUBLISHEDを生成") {
            val status = BookPublishStatus.of(1)
            status shouldBe BookPublishStatus.UNPUBLISHED
        }

        test("値2からPUBLISHEDを生成") {
            val status = BookPublishStatus.of(2)
            status shouldBe BookPublishStatus.PUBLISHED
        }

        test("存在しない値でNoSuchElementExceptionが発生") {
            shouldThrow<NoSuchElementException> {
                BookPublishStatus.of(999)
            }
        }
    }

    context("canChangeメソッドのテスト") {
        test("UNPUBLISHEDからUNPUBLISHEDへの変更は可能") {
            BookPublishStatus.UNPUBLISHED.canChange(BookPublishStatus.UNPUBLISHED) shouldBe true
        }

        test("UNPUBLISHEDからPUBLISHEDへの変更は可能") {
            BookPublishStatus.UNPUBLISHED.canChange(BookPublishStatus.PUBLISHED) shouldBe true
        }

        test("PUBLISHEDからPUBLISHEDへの変更は可能") {
            BookPublishStatus.PUBLISHED.canChange(BookPublishStatus.PUBLISHED) shouldBe true
        }

        test("PUBLISHEDからUNPUBLISHEDへの変更は不可能") {
            BookPublishStatus.PUBLISHED.canChange(BookPublishStatus.UNPUBLISHED) shouldBe false
        }
    }

    context("値の確認") {
        test("UNPUBLISHEDの値は1") {
            BookPublishStatus.UNPUBLISHED.value shouldBe 1
        }

        test("PUBLISHEDの値は2") {
            BookPublishStatus.PUBLISHED.value shouldBe 2
        }
    }
})
