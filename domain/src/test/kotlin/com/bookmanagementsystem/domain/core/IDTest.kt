package com.bookmanagementsystem.domain.core

import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe

class IDTest : FunSpec({
    test("有効な正の値でIDを作成できる") {
        val id = ID<String>(1)
        id.value shouldBe 1
    }

    test("0でIDを作成する際は例外が発生する") {
        shouldThrow<IllegalArgumentException> {
            ID<String>(0)
        }
    }

    test("負の値でIDを作成する際は例外が発生する") {
        shouldThrow<IllegalArgumentException> {
            ID<String>(-1)
        }
    }
})
