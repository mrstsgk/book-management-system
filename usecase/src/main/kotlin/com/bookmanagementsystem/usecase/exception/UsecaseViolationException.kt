package com.bookmanagementsystem.usecase.exception

/**
 * ユースケース層でのバリデーション違反を表す例外
 */
class UsecaseViolationException(
    override val message: String
) : RuntimeException(message)
