package com.bookmanagementsystem.usecase.validation.core

import com.bookmanagementsystem.domain.core.ID
import jakarta.validation.Constraint
import jakarta.validation.ConstraintValidator
import jakarta.validation.ConstraintValidatorContext
import jakarta.validation.Payload
import kotlin.reflect.KClass

/**
 * リスト内の要素が重複していないことを検証するアノテーション
 */
@Target(AnnotationTarget.FIELD, AnnotationTarget.PROPERTY_GETTER)
@Retention(AnnotationRetention.RUNTIME)
@Constraint(validatedBy = [UniqueIdValidator::class])
@MustBeDocumented
annotation class UniqueId(
    val message: String = "リスト内に重複した要素があります",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

class UniqueIdValidator : ConstraintValidator<UniqueId, List<ID<*>>> {
    /**
     * 配列内のIDが重複していないかを検証する
     */
    override fun isValid(value: List<ID<*>>, context: ConstraintValidatorContext?): Boolean {
        // リストのサイズと重複を除いたSetのサイズが同じかチェック
        return value.size == value.toSet().size
    }
}
