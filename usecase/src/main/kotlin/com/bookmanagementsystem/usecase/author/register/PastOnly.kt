package com.bookmanagementsystem.usecase.author.register

import jakarta.validation.Constraint
import jakarta.validation.ConstraintValidator
import jakarta.validation.ConstraintValidatorContext
import jakarta.validation.Payload
import java.time.LocalDate
import kotlin.reflect.KClass

/**
 * 現在の日付よりも過去であることを検証するアノテーション
 * 当日は含まない
 */
@Target(AnnotationTarget.FIELD, AnnotationTarget.PROPERTY_GETTER)
@Retention(AnnotationRetention.RUNTIME)
@Constraint(validatedBy = [PastOnlyValidator::class])
@MustBeDocumented
annotation class PastOnly(
    val message: String = "生年月日は現在の日付よりも過去である必要があります",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

/**
 * PastOnlyアノテーションのバリデーター
 */
class PastOnlyValidator : ConstraintValidator<PastOnly, LocalDate?> {

    override fun isValid(value: LocalDate?, context: ConstraintValidatorContext?): Boolean {
        if (value == null) {
            return true // nullは他のバリデーションで処理
        }
        return value.isBefore(LocalDate.now())
    }
}
