package com.bookmanagementsystem.usecase.validation

import jakarta.validation.Validator
import org.springframework.stereotype.Component

/**
 * コマンドオブジェクトのバリデーションを行うバリデーター
 */
@Component
class CommandValidator(
    private val validator: Validator
) {
    /**
     * コマンドオブジェクトをバリデーションする
     * @param command バリデーション対象のコマンド
     * @return バリデーションエラーメッセージのリスト
     */
    fun <T> validate(command: T): List<String> {
        val violations = validator.validate(command)
        return violations.map { violation ->
            "${violation.propertyPath}: ${violation.message}"
        }
    }
}
