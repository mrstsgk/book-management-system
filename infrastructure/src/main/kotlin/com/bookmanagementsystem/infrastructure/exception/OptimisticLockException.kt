package com.bookmanagementsystem.infrastructure.exception

class OptimisticLockException(
    message: String,
    cause: Throwable? = null
) : RuntimeException(message, cause)