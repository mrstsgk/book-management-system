package com.bookmanagementsystem.presentation.controller.author

import com.bookmanagementsystem.config.IntegrationTestWithSql
import com.bookmanagementsystem.presentation.model.AuthorResponseModel
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

@IntegrationTestWithSql(sqlScript = "CreateAuthorControllerTest.sql")
class CreateAuthorControllerTest : FunSpec() {
    override fun extensions() = listOf(SpringExtension)

    @Autowired
    private lateinit var webApplicationContext: WebApplicationContext

    @Autowired
    private lateinit var objectMapper: ObjectMapper
    private lateinit var mockMvc: MockMvc

    init {
        context("200") {
            test("有効なリクエストで著者が正常に作成される") {
                // Arrange
                mockMvc = MockMvcBuilders
                    .webAppContextSetup(webApplicationContext)
                    .build()

                val request = CreateAuthorRequestModel(
                    name = "夏目漱石",
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)

                // Act & Assert
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

            test("生年月日がnullでも著者が正常に作成される") {
                // Arrange
                mockMvc = MockMvcBuilders
                    .webAppContextSetup(webApplicationContext)
                    .build()

                val request = CreateAuthorRequestModel(
                    name = "太宰治",
                    birthDate = null
                )
                val requestJson = objectMapper.writeValueAsString(request)

                // Act & Assert
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
                response.name shouldBe "太宰治"
                response.birthDate shouldBe null
            }
        }

        context("400") {
            test("nameがnullの場合バリデーションエラーが発生する") {
                // Arrange
                mockMvc = MockMvcBuilders
                    .webAppContextSetup(webApplicationContext)
                    .build()

                val request = CreateAuthorRequestModel(
                    name = null,
                    birthDate = LocalDate.of(1867, 2, 9)
                )
                val requestJson = objectMapper.writeValueAsString(request)

                // Act & Assert
                mockMvc.perform(
                    post("/api/authors")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
            }

            test("空のJSONの場合バリデーションエラーが発生する") {
                // Arrange
                mockMvc = MockMvcBuilders
                    .webAppContextSetup(webApplicationContext)
                    .build()

                val requestJson = "{}"

                // Act & Assert
                mockMvc.perform(
                    post("/api/authors")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(requestJson)
                )
                    .andExpect(status().isBadRequest)
            }
        }
    }
}
