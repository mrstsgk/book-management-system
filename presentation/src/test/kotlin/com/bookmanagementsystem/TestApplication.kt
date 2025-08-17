package com.bookmanagementsystem

import org.springframework.boot.autoconfigure.SpringBootApplication

@SpringBootApplication(
    scanBasePackages = [
        "com.bookmanagementsystem.presentation",
        "com.bookmanagementsystem.infrastructure",
        "com.bookmanagementsystem.usecase"
    ]
)
class TestApplication
