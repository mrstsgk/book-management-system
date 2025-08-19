package com.bookmanagementsystem.presentation.controller.book

import com.bookmanagementsystem.presentation.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.BadRequestErrorResponseModel
import com.bookmanagementsystem.presentation.model.BookResponseModel
import com.bookmanagementsystem.presentation.model.BookStatus
import com.bookmanagementsystem.presentation.model.NotFoundErrorResponseModel
import com.bookmanagementsystem.presentation.model.OptimisticLockErrorResponseModel
import com.bookmanagementsystem.presentation.model.UpdateBookRequestModel
import com.fasterxml.jackson.databind.ObjectMapper
import io.kotest.core.spec.style.FunSpec
import io.kotest.extensions.spring.SpringExtension
import io.kotest.matchers.shouldBe
import io.kotest.matchers.shouldNotBe
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.content
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.status
import org.springframework.test.web.servlet.setup.MockMvcBuilders
import org.springframework.web.context.WebApplicationContext
import java.time.LocalDate

/**
 * UpdateBookController の統合テスト。
 *
 * 実際に API を通じて書籍更新を行い、バリデーションや正常系の挙動を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + MockMvc）で実行される。
 *
 * ### 各HTTP ステータス毎1パス通せばOK
 *  context("200") - 正常系：
 *  context("400") - 異常系：
 *  context("404") - 存在しない書籍：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * 異常系・正常系ともに他のテストでカバレッジを確保しており、基本的な入力妥当性を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "UpdateBookControllerTest.sql")
class UpdateBookControllerTest : FunSpec() {
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
            test("有効なリクエストで書籍が正常に更新される") {
                val request = UpdateBookRequestModel(
                    title = "更新された吾輩は猫である",
                    price = 2000L,
                    authorIds = listOf(1),
                    status = BookStatus.PUBLISHED,
                    version = 1
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    put("/api/books/1")
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
                response.title shouldBe "更新された吾輩は猫である"
                response.price shouldBe 2000L
                response.authors?.size shouldBe 1
                response.authors?.first()?.name shouldBe "夏目漱石"
                response.authors?.first()?.birthDate shouldBe LocalDate.of(1867, 2, 9)
                response.status shouldBe BookStatus.PUBLISHED
            }

            test("複数の著者を持つ書籍が正常に更新される") {
                val request = UpdateBookRequestModel(
                    title = "更新された日本文学選集",
                    price = 3500L,
                    authorIds = listOf(1, 2, 3),
                    status = BookStatus.PUBLISHED,
                    version = 1
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    put("/api/books/2")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response = objectMapper.readValue(responseJson, BookResponseModel::class.java)

                response.id shouldBe 2
                response.title shouldBe "更新された日本文学選集"
                response.price shouldBe 3500L
                response.authors?.size shouldBe 3
                response.authors?.get(0)?.id shouldBe 1
                response.authors?.get(0)?.name shouldBe "夏目漱石"
                response.authors?.get(0)?.birthDate shouldBe LocalDate.of(1867, 2, 9)
                response.authors?.get(0)?.version shouldBe 1
                response.authors?.get(1)?.id shouldBe 2
                response.authors?.get(1)?.name shouldBe "太宰治"
                response.authors?.get(1)?.birthDate shouldBe LocalDate.of(1909, 6, 19)
                response.authors?.get(1)?.version shouldBe 1
                response.authors?.get(2)?.id shouldBe 3
                response.authors?.get(2)?.name shouldBe "芥川龍之介"
                response.authors?.get(2)?.birthDate shouldBe LocalDate.of(1892, 3, 1)
                response.authors?.get(2)?.version shouldBe 1
                response.status shouldBe BookStatus.PUBLISHED
            }
        }

        context("400") {
            test("価格が負の値ならばバリデーションエラーが発生する") {
                val request = UpdateBookRequestModel(
                    title = "負の価格の書籍",
                    price = -100L,
                    authorIds = listOf(1),
                    status = BookStatus.UNPUBLISHED,
                    version = 1
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/books/2") // 書籍ID 2は未出版(UNPUBLISHED)状態
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // BadRequestErrorResponseModel形式のレスポンス検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, BadRequestErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "VALIDATION_ERROR"
                errorResponse.errors?.first()?.message?.contains("price") shouldBe true
            }

            test("出版済みから未出版への変更でバリデーションエラーが発生する") {
                val request = UpdateBookRequestModel(
                    title = "出版状況変更テスト",
                    price = 1000L,
                    authorIds = listOf(1),
                    status = BookStatus.UNPUBLISHED,
                    version = 1
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/books/1") // 書籍ID 1は出版済み(PUBLISHED)状態
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // BadRequestErrorResponseModel形式のレスポンス検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, BadRequestErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "VALIDATION_ERROR"
                errorResponse.errors?.first()?.message shouldBe ": 出版状況は「出版済み」から「未出版」に変更できません"
            }
        }

        context("409") {
            test("バージョンが古い場合楽観的ロック競合エラーが発生する") {
                val request = UpdateBookRequestModel(
                    title = "競合テスト用書籍",
                    price = 1500L,
                    authorIds = listOf(1),
                    status = BookStatus.PUBLISHED,
                    version = 0 // 古いバージョン（現在のデータはversion=1）
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/books/1")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isConflict)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // OptimisticLockErrorResponseModel形式のレスポンス検証（OpenAPI準拠）
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, OptimisticLockErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "OPTIMISTIC_LOCK_EXCEPTION"
                errorResponse.errors?.first()?.message shouldBe "書籍の更新に失敗しました。"
            }
        }

        context("404") {
            test("存在しない書籍IDを指定した場合NotFoundErrorResponseModelが返される") {
                val request = UpdateBookRequestModel(
                    title = "存在しない書籍",
                    price = 1000L,
                    authorIds = listOf(1),
                    status = BookStatus.PUBLISHED,
                    version = 1
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/books/999")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isNotFound)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // NotFoundErrorResponseModel形式のレスポンス検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, NotFoundErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "NOT_FOUND"
                errorResponse.errors?.first()?.message shouldNotBe null
            }
        }
    }
}
