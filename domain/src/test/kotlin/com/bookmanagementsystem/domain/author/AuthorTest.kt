package com.bookmanagementsystem.domain.author

import com.bookmanagementsystem.domain.core.ID
import io.kotest.core.spec.style.FunSpec
import io.kotest.matchers.shouldBe
import java.time.LocalDate

class AuthorTest : FunSpec({
    test("有効なパラメータで著者を作成できる") {
        val authorId = ID<Author>(1)
        val name = "太宰治"
        val authorBirthDate = AuthorBirthDate(LocalDate.of(1909, 6, 19))

        val author = Author(authorId, name, authorBirthDate)

        author.id shouldBe authorId
        author.name shouldBe name
        author.birthDate shouldBe authorBirthDate
    }
})
