package com.bookmanagementsystem.usecase.exception

/**
 * ユースケース層でのバリデーション違反を表す例外
 */
class UsecaseViolationException(
    val errors: List<String>
) : RuntimeException(errors.joinToString(", ")) {
    constructor(message: String) : this(listOf(message))
}
