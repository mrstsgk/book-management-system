package com.bookmanagementsystem.infrastructure.author

import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.book.Book
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.references.AUTHOR
import com.bookmanagementsystem.jooq.tables.references.AUTHOR_BOOK
import com.bookmanagementsystem.usecase.author.AuthorDto
import com.bookmanagementsystem.usecase.author.read.AuthorQueryService
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class AuthorQueryServiceImpl(private val dsl: DSLContext) : AuthorQueryService {
    override fun findByBookId(bookId: ID<Book>) = dsl
        .select(
            AUTHOR.ID,
            AUTHOR.NAME,
            AUTHOR.BIRTH_DATE,
            AUTHOR.VERSION,
        )
        .from(AUTHOR)
        .join(AUTHOR_BOOK)
        .on(AUTHOR.ID.eq(AUTHOR_BOOK.AUTHOR_ID))
        .where(AUTHOR_BOOK.BOOK_ID.eq(bookId.value))
        .orderBy(AUTHOR.ID)
        .fetch()
        .map { record ->
            AuthorDto(
                id = ID(record[AUTHOR.ID]!!),
                name = record[AUTHOR.NAME]!!,
                birthDate = record[AUTHOR.BIRTH_DATE]?.let { AuthorBirthDate(it) },
                version = record[AUTHOR.VERSION]!!
            )
        }
}
