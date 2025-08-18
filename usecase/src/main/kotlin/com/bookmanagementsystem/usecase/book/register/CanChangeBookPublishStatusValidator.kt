package com.bookmanagementsystem.usecase.book.register

import com.bookmanagementsystem.domain.book.BookRepository
import jakarta.validation.Constraint
import jakarta.validation.ConstraintValidator
import jakarta.validation.ConstraintValidatorContext
import jakarta.validation.Payload
import kotlin.reflect.KClass

/**
 * 出版済みステータスのものを未出版には変更できないことを検証するアノテーション
 */
@Target(AnnotationTarget.CLASS)
@Retention(AnnotationRetention.RUNTIME)
@Constraint(validatedBy = [CanChangeBookPublishStatusValidator::class])
@MustBeDocumented
annotation class CanChangeBookPublishStatus(
    val message: String = "出版状況は「出版済み」から「未出版」に変更できません",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

class CanChangeBookPublishStatusValidator(
    private val repository: BookRepository
) : ConstraintValidator<CanChangeBookPublishStatus, UpdateBookCommand> {

    override fun isValid(value: UpdateBookCommand, context: ConstraintValidatorContext?): Boolean {
        val currentStatus = repository.getBookPublishStatusById(value.id)
        return currentStatus.canChange(value.status)
    }
}
