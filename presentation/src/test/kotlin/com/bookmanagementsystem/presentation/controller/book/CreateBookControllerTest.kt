package com.bookmanagementsystem.presentation.controller.book

import com.bookmanagementsystem.presentation.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.BadRequestErrorResponseModel
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.presentation.model.CreateBookRequestModel
import com.fasterxml.jackson.databind.ObjectMapper
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.content
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.status
import org.springframework.test.web.servlet.setup.MockMvcBuilders
import org.springframework.web.context.WebApplicationContext
import java.time.LocalDate

/**
 * CreateBookController の統合テスト。
 *
 * 実際に API を通じて書籍登録を行い、バリデーションや正常系の挙動を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + MockMvc）で実行される。
 *
 * ### 各HTTP ステータスで1パス通せばOK
 *  context("200") - 正常系：
 *  context("400") - 異常系：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * 異常系・正常系ともに他のテストでカバレッジを確保しており、基本的な入力妥当性を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "CreateBookControllerTest.sql")
class CreateBookControllerTest : FunSpec() {
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
            test("有効なリクエストで書籍が正常に作成される") {
                val request = CreateBookRequestModel(
                    title = "吾輩は猫である",
                    price = 1500L,
                    authorIds = listOf(1),
                    status = BookStatus.PUBLISHED
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    post("/api/books")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response = objectMapper.readValue(responseJson, BookResponseModel::class.java)

                response.id shouldBe 1
                response.title shouldBe "吾輩は猫である"
                response.price shouldBe 1500L
                response.authors?.size shouldBe 1
                response.authors?.first()?.name shouldBe "夏目漱石"
                response.authors?.first()?.birthDate shouldBe LocalDate.of(1867, 2, 9)
                response.status shouldBe BookStatus.PUBLISHED
            }

            test("複数の著者を持つ書籍が正常に作成される") {
                val request = CreateBookRequestModel(
                    title = "日本文学選集",
                    price = 3000L,
                    authorIds = listOf(1, 2, 3),
                    status = BookStatus.PUBLISHED
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    post("/api/books")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response = objectMapper.readValue(responseJson, BookResponseModel::class.java)

                response.title shouldBe "日本文学選集"
                response.price shouldBe 3000L
                response.authors?.size shouldBe 3

                // 著者が正しく取得されていることを確認
                val authorNames = response.authors?.map { it.name } ?: emptyList()
                authorNames shouldBe listOf("夏目漱石", "太宰治", "芥川龍之介")

                // 各著者の詳細情報も確認
                response.authors?.get(0)?.id shouldBe 1
                response.authors?.get(0)?.name shouldBe "夏目漱石"
                response.authors?.get(0)?.birthDate shouldBe LocalDate.of(1867, 2, 9)

                response.authors?.get(1)?.id shouldBe 2
                response.authors?.get(1)?.name shouldBe "太宰治"
                response.authors?.get(1)?.birthDate shouldBe LocalDate.of(1909, 6, 19)

                response.authors?.get(2)?.id shouldBe 3
                response.authors?.get(2)?.name shouldBe "芥川龍之介"
                response.authors?.get(2)?.birthDate shouldBe LocalDate.of(1892, 3, 1)

                response.status shouldBe BookStatus.PUBLISHED
            }
        }

        context("400") {
            test("空のJSONの場合バリデーションエラーが発生する") {
                val requestJson = "{}"

                val result = mockMvc.perform(
                    post("/api/books")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // BadRequestErrorResponseModel形式のレスポンス検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, BadRequestErrorResponseModel::class.java)

                errorResponse.errors?.isNotEmpty() shouldBe true
                val errorMessages = errorResponse.errors?.map { it.message } ?: emptyList()
                errorMessages.toSet() shouldBe setOf(
                    "title: must not be null",
                    "price: must not be null",
                    "authorIds: must not be null",
                    "status: must not be null"
                )
            }
        }
    }
}
