package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.presentation.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.BadRequestErrorResponseModel
import com.bookmanagementsystem.presentation.model.NotFoundErrorResponseModel
import com.bookmanagementsystem.presentation.model.OptimisticLockErrorResponseModel
import com.bookmanagementsystem.presentation.model.UpdateAuthorRequestModel
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
 * UpdateAuthorController の統合テスト。
 *
 * 実際に API を通じて著者更新を行い、バリデーションや正常系の挙動を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + MockMvc）で実行される。
 *
 * ### 各HTTP ステータスで1パス通せばOK
 *  context("200") - 正常系：
 *  context("400") - 異常系：
 *  context("404") - 存在しない著者：
 *  context("409") - 楽観的ロックエラー：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * 異常系・正常系ともに他のテストでカバレッジを確保しており、基本的な入力妥当性を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "UpdateAuthorControllerTest.sql")
class UpdateAuthorControllerTest : FunSpec() {
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
            test("有効なリクエストで著者が正常に更新される") {
                val request = UpdateAuthorRequestModel(
                    name = "更新された夏目漱石",
                    version = 1,
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    put("/api/authors/1")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isOk)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // レスポンスボディの検証
                val responseJson = result.response.contentAsString
                val response = objectMapper.readValue(responseJson, AuthorResponseModel::class.java)

                response.id shouldBe 1
                response.name shouldBe "更新された夏目漱石"
                response.birthDate shouldBe LocalDate.of(1867, 2, 9)
                response.version shouldBe 2
            }
        }

        context("400") {
            test("生年月日が未来日付ならばバリデーションエラーが発生する（JSR-303）") {
                val request = UpdateAuthorRequestModel(
                    name = "未来の著者",
                    version = 1,
                    birthDate = LocalDate.now().plusDays(1) // 未来日付
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/authors/1")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // BadRequestErrorResponseModel形式のレスポンス検証（OpenAPI準拠）
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, BadRequestErrorResponseModel::class.java)

                errorResponse.errors?.size shouldBe 1
                errorResponse.errors?.first()?.code shouldBe "VALIDATION_ERROR"
                errorResponse.errors?.first()?.message shouldBe "birthDate: 生年月日は現在の日付よりも過去である必要があります"
            }
        }

        context("404") {
            test("存在しない著者IDを指定した場合NotFoundErrorResponseModelが返される") {
                val request = UpdateAuthorRequestModel(
                    name = "存在しない著者",
                    version = 1,
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/authors/999")
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

        context("409") {
            test("バージョンが古い場合楽観的ロック競合エラーが発生する") {
                val request = UpdateAuthorRequestModel(
                    name = "競合テスト用著者",
                    version = 0, // 古いバージョン（現在のデータはversion=1）
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)

                val result = mockMvc.perform(
                    put("/api/authors/1")
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
                errorResponse.errors?.first()?.message shouldBe "著者の更新に失敗しました。"
            }
        }
    }
}
