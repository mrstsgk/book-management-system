package com.bookmanagementsystem.usecase.author

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import jakarta.validation.Constraint
import jakarta.validation.ConstraintValidator
import jakarta.validation.ConstraintValidatorContext
import jakarta.validation.Payload
import kotlin.reflect.KClass

@Target(AnnotationTarget.FIELD, AnnotationTarget.PROPERTY_GETTER)
@Retention(AnnotationRetention.RUNTIME)
@Constraint(validatedBy = [AuthorExistsValidator::class])
@MustBeDocumented
annotation class AuthorExists(
    val message: String = "存在しない著者が指定されています",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

/**
 * 著者が存在するかどうかを検証するバリデータ
 */
class AuthorExistsValidator(
    private val repository: AuthorRepository
) : ConstraintValidator<AuthorExists, List<ID<Author>>> {

    override fun isValid(value: List<ID<Author>>, context: ConstraintValidatorContext?): Boolean {
        return value.isEmpty() || existsAll(value.distinct())
    }

    private fun existsAll(ids: List<ID<Author>>): Boolean {
        val authors = repository.findByIds(ids)
        return authors.size == ids.size
    }
}
