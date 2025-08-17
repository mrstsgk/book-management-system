package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.presentation.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
import com.bookmanagementsystem.presentation.model.BadRequestErrorResponseModel
import com.bookmanagementsystem.presentation.model.CreateAuthorRequestModel
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
 * CreateAuthorController の統合テスト。
 *
 * 実際に API を通じて著者登録を行い、バリデーションや正常系の挙動を検証する。
 * モックではなく実際の環境に近い構成（Spring Boot + MockMvc）で実行される。
 *
 * ### 各HTTP ステータスで1パス通せばOK
 *  context("200") - 正常系：
 *  context("400") - 異常系：
 *
 * 本テストでは、最低限「1パス通ること」で動作確認済みとする。
 * 異常系・正常系ともに他のテストでカバレッジを確保しており、基本的な入力妥当性を保証する目的。
 */
@IntegrationTestWithSql(sqlScript = "CreateAuthorControllerTest.sql")
class CreateAuthorControllerTest : FunSpec() {
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
            test("有効なリクエストで著者が正常に作成される") {
                val request = CreateAuthorRequestModel(
                    name = "夏目漱石",
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)
                val result = mockMvc.perform(
                    post("/api/authors")
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
                response.name shouldBe "夏目漱石"
                response.birthDate shouldBe LocalDate.of(1867, 2, 9)
            }
        }

        context("400") {
            test("空のJSONの場合バリデーションエラーが発生する") {
                val requestJson = "{}"

                val result = mockMvc.perform(
                    post("/api/authors")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
                    .andExpect(content().contentType(MediaType.APPLICATION_JSON))
                    .andReturn()

                // BadRequestErrorResponseModel形式のレスポンス検証
                val responseJson = result.response.contentAsString
                val errorResponse = objectMapper.readValue(responseJson, BadRequestErrorResponseModel::class.java)

                errorResponse.errors!!.size shouldBe 1
                errorResponse.errors.first().code shouldBe "VALIDATION_ERROR"
                errorResponse.errors.first().message shouldBe "name: must not be null"
            }
        }
    }
}
