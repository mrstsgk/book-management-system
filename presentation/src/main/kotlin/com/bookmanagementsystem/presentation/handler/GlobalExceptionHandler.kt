package com.bookmanagementsystem.presentation.handler

import com.bookmanagementsystem.presentation.model.BadRequestErrorResponseModel
import com.bookmanagementsystem.presentation.model.ErrorModel
import com.bookmanagementsystem.presentation.model.NotFoundErrorResponseModel
import com.bookmanagementsystem.usecase.exception.UsecaseViolationException
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.MethodArgumentNotValidException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice

@RestControllerAdvice
class GlobalExceptionHandler {
    @ExceptionHandler(MethodArgumentNotValidException::class)
    fun handleValidationException(ex: MethodArgumentNotValidException): ResponseEntity<BadRequestErrorResponseModel> {
        val errors = ex.bindingResult.fieldErrors.map { fieldError ->
            ErrorModel(
                code = "VALIDATION_ERROR",
                message = "${fieldError.field}: ${fieldError.defaultMessage}"
            )
        }

        val errorResponse = BadRequestErrorResponseModel(errors = errors)
        return ResponseEntity(errorResponse, HttpStatus.BAD_REQUEST)
    }

    @ExceptionHandler(UsecaseViolationException::class)
    fun handleUsecaseViolationException(ex: UsecaseViolationException): ResponseEntity<BadRequestErrorResponseModel> {
        val errors = ex.errors.map { error ->
            ErrorModel(
                code = "VALIDATION_ERROR",
                message = error
            )
        }
        val errorResponse = BadRequestErrorResponseModel(errors = errors)
        return ResponseEntity(errorResponse, HttpStatus.BAD_REQUEST)
    }

    @ExceptionHandler(NoSuchElementException::class)
    fun handleNotFound(e: NoSuchElementException): ResponseEntity<NotFoundErrorResponseModel> {
        val errorModel = ErrorModel(
            code = "NOT_FOUND",
            message = e.message ?: "リソースが見つかりません"
        )
        val error = NotFoundErrorResponseModel(
            errors = listOf(errorModel)
        )
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error)
    }

    @ExceptionHandler(IllegalArgumentException::class)
    fun handleIllegalArgumentException(ex: IllegalArgumentException): ResponseEntity<Map<String, String>> {
        val errorResponse = mapOf("error" to (ex.message ?: "不正な引数です"))
        return ResponseEntity(errorResponse, HttpStatus.INTERNAL_SERVER_ERROR)
    }
}
