package com.bookmanagementsystem

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication(
    scanBasePackages = ["com.bookmanagementsystem.presentation", "com.bookmanagementsystem.infrastructure"]
)
class BookmanagementsystemApplication

fun main(args: Array<String>) {
    @Suppress("SpreadOperator") // NOTE: 許容する（runApplicationに引数を渡すためにスプレッド演算子（*args）を使用している）
    runApplication<BookmanagementsystemApplication>(*args)
}
