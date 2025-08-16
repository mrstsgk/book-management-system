package com.bookmanagementsystem.domain.book

import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe

class PublishStatusTest : FunSpec({
    test("未出版ステータスは値1を持つ") {
        BookPublishStatus.UNPUBLISHED.value shouldBe 1
    }

    test("出版済みステータスは値2を持つ") {
        BookPublishStatus.PUBLISHED.value shouldBe 2
    }

    test("値1から未出版ステータスを作成できる") {
        BookPublishStatus.of(1) shouldBe BookPublishStatus.UNPUBLISHED
    }

    test("値2から出版済みステータスを作成できる") {
        BookPublishStatus.of(2) shouldBe BookPublishStatus.PUBLISHED
    }

    test("不正な値から出版ステータスを作成できない") {
        shouldThrow<NoSuchElementException> {
            BookPublishStatus.of(0)
        }
    }

    test("未出版は出版可能") {
        val status = BookPublishStatus.UNPUBLISHED
        status.canPublish() shouldBe true
    }

    test("出版済みは出版不可能") {
        val status = BookPublishStatus.PUBLISHED
        status.canPublish() shouldBe false
    }
})
