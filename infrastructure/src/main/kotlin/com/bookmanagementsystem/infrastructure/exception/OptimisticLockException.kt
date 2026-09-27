package com.bookmanagementsystem.infrastructure.exception

/**
 * 楽観的ロック例外
 */
class OptimisticLockException(message: String) : RuntimeException(message)
