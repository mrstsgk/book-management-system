package com.bookmanagementsystem.infrastructure.author

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.jooq.tables.Author.Companion.AUTHOR
import com.bookmanagementsystem.jooq.tables.records.AuthorRecord
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class AuthorRepositoryImpl(private val dsl: DSLContext) : AuthorRepository {
    override fun insert(author: Author) = dsl
        .insertInto(AUTHOR)
        .set(AUTHOR.NAME, author.name)
        .set(AUTHOR.BIRTH_DATE, author.birthDate?.value)
        .returning()
        .fetchSingle()
        .let { convert(it) }

    private fun convert(record: AuthorRecord) = Author(
        id = ID(record.id!!),
        name = record.name!!,
        birthDate = record.birthDate?.let { date -> AuthorBirthDate(date) }
    )
}
