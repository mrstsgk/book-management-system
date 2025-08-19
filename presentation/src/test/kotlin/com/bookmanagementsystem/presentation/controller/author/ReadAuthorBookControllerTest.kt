package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.presentation.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.presentation.model.NotFoundErrorResponseModel
import com.fasterxml.jackson.core.type.TypeReference
import com.fasterxml.jackson.databind.ObjectMapper
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.content
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.status
import org.springframework.test.web.servlet.setup.MockMvcBuilders
import org.springframework.web.context.WebApplicationContext
import java.time.LocalDate

/**
 * ReadAuthorBookController の統合テスト。
 *
 * 実際に API を通じて著者に紐づく書籍一覧の取得を行い、正常系・異常系の挙動を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + MockMvc）で実行される。
 *
 * ### 各HTTP ステータスで1パス通せばOK
 *  context("200") - 正常系：
 *  context("404") - 異常系：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * 異常系・正常系ともに他のテストでカバレッジを確保しており、基本的な入力妥当性を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "ReadAuthorBookControllerTest.sql")
class ReadAuthorBookControllerTest : FunSpec() {
    override fun extensions() = listOf(SpringExtension)

    @Autowired
    private lateinit var webApplicationContext: WebApplicationContext

    @Autowired
    private lateinit var objectMapper: ObjectMapper
    private lateinit var mockMvc: MockMvc

    init {
        beforeEach {
            mockMvc = MockMvcBuilders
                .webAppContextSetup(webApplicationContext)
                .build()
        }

        context("200") {
            test("著者IDに紐づく書籍一覧が正常に取得される") {
                val result = mockMvc.perform(
                    get("/api/authors/1/books")
                        .contentType(MediaType.APPLICATION_JSON)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response = objectMapper.readValue(
                    responseJson,
                    object : TypeReference<List<BookResponseModel>>() {}
                )

                response.size shouldBe 2

                // 1冊目の書籍検証
                val book1 = response[0]
                book1.id shouldBe 1
                book1.title shouldBe "吾輩は猫である"
                book1.price shouldBe 1500L
                book1.status shouldBe BookStatus.PUBLISHED
                book1.version shouldBe 1
                book1.authors?.size shouldBe 1
                book1.authors?.get(0)?.id shouldBe 1
                book1.authors?.get(0)?.name shouldBe "夏目漱石"
                book1.authors?.get(0)?.birthDate shouldBe LocalDate.of(1867, 2, 9)
                book1.authors?.get(0)?.version shouldBe 1

                // 2冊目の書籍検証（共著）
                val book2 = response[1]
                book2.id shouldBe 2
                book2.title shouldBe "日本文学選集"
                book2.price shouldBe 3000L
                book2.status shouldBe BookStatus.PUBLISHED
                book2.version shouldBe 1
                book2.authors?.size shouldBe 3
            }

            test("著者に紐づく書籍が存在しない場合、空の配列が返される") {
                val result = mockMvc.perform(
                    get("/api/authors/5/books")
                        .contentType(MediaType.APPLICATION_JSON)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response =
                    objectMapper.readValue(responseJson, object : TypeReference<List<BookResponseModel>>() {})

                response.size shouldBe 0
            }

            test("単独著作の書籍が正常に取得される") {
                val result = mockMvc.perform(
                    get("/api/authors/2/books")
                        .contentType(MediaType.APPLICATION_JSON)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response =
                    objectMapper.readValue(responseJson, object : TypeReference<List<BookResponseModel>>() {})

                response.size shouldBe 2

                // 単独著作の書籍検証
                val singleAuthorBook = response.find { it.id == 4 }
                singleAuthorBook!!.title shouldBe "文学論"
                singleAuthorBook.price shouldBe 2500L
                singleAuthorBook.version shouldBe 1
                singleAuthorBook.authors?.size shouldBe 1
                singleAuthorBook.authors?.get(0)?.id shouldBe 2
                singleAuthorBook.authors?.get(0)?.name shouldBe "太宰治"
                singleAuthorBook.authors?.get(0)?.version shouldBe 1
            }
        }

        context("404") {
            test("存在しない著者IDの場合、404エラーが発生する") {
                val result = mockMvc.perform(
                    get("/api/authors/999/books")
                        .contentType(MediaType.APPLICATION_JSON)
                )
                    .andExpect(status().isNotFound)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // エラーレスポンスボディの検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, NotFoundErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "NOT_FOUND"
                errorResponse.errors?.first()?.message shouldBe "著者が見つかりません。ID: 999"
            }
        }
    }
}
