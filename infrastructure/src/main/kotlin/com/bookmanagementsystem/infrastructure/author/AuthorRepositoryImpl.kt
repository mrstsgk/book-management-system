package com.bookmanagementsystem.infrastructure.author

import com.bookmanagementsystem.domain.author.Author
import com.bookmanagementsystem.domain.author.AuthorBirthDate
import com.bookmanagementsystem.domain.author.AuthorRepository
import com.bookmanagementsystem.domain.core.ID
import com.bookmanagementsystem.infrastructure.exception.OptimisticLockException
import com.bookmanagementsystem.jooq.tables.Author.Companion.AUTHOR
import com.bookmanagementsystem.jooq.tables.records.AuthorRecord
import org.jooq.DSLContext
import org.springframework.stereotype.Repository

@Repository
class AuthorRepositoryImpl(private val dsl: DSLContext) : AuthorRepository {
    override fun findById(id: ID<Author>): Author? = dsl
        .selectFrom(AUTHOR)
        .where(AUTHOR.ID.eq(id.value))
        .fetchOne()
        ?.let { convert(it) }

    override fun findByIds(ids: List<ID<Author>>): List<Author> = dsl
        .selectFrom(AUTHOR)
        .where(AUTHOR.ID.`in`(ids.map { it.value }))
        .fetch()
        .map { convert(it) }

    override fun insert(author: Author) = dsl
        .insertInto(AUTHOR)
        .set(AUTHOR.NAME, author.name)
        .set(AUTHOR.BIRTH_DATE, author.birthDate?.value)
        .set(AUTHOR.VERSION, 1)
        .returning()
        .fetchSingle()
        .let { convert(it) }

    override fun update(author: Author): Author {
        val record = dsl
            .update(AUTHOR)
            .set(AUTHOR.NAME, author.name)
            .set(AUTHOR.BIRTH_DATE, author.birthDate?.value)
            .set(AUTHOR.VERSION, author.version!! + 1)
            .where(AUTHOR.ID.eq(author.id!!.value))
            .and(AUTHOR.VERSION.eq(author.version))
            .returning()
            .fetchOne()
            ?: throw OptimisticLockException("著者の更新に失敗しました。")

        return convert(record)
    }

    private fun convert(record: AuthorRecord) = Author(
        id = ID(record.id!!),
        name = record.name!!,
        birthDate = record.birthDate?.let { date -> AuthorBirthDate(date) },
        version = record.version
    )
}
