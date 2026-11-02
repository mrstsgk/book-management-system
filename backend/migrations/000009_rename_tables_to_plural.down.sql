ALTER TABLE sessions RENAME CONSTRAINT sessions_pkey TO admin_session_pkey;
ALTER TABLE sessions RENAME TO admin_session;

ALTER TABLE book_tags RENAME CONSTRAINT book_tags_tag_id_fkey TO book_tag_tag_id_fkey;
ALTER TABLE book_tags RENAME CONSTRAINT book_tags_book_id_fkey TO book_tag_book_id_fkey;
ALTER TABLE book_tags RENAME CONSTRAINT book_tags_pkey TO book_tag_pkey;
ALTER TABLE book_tags RENAME TO book_tag;

ALTER SEQUENCE tags_id_seq RENAME TO tag_id_seq;
ALTER TABLE tags RENAME CONSTRAINT uq_tags_name TO uq_tag_name;
ALTER TABLE tags RENAME CONSTRAINT tags_pkey TO tag_pkey;
ALTER TABLE tags RENAME TO tag;

ALTER SEQUENCE books_id_seq RENAME TO book_id_seq;
ALTER TABLE books RENAME CONSTRAINT ck_books_cover_pair TO ck_book_cover_pair;
ALTER TABLE books RENAME CONSTRAINT ck_books_cover_source TO ck_book_cover_source;
ALTER TABLE books RENAME CONSTRAINT ck_books_rating TO ck_book_rating;
ALTER TABLE books RENAME CONSTRAINT uq_books_isbn TO uq_book_isbn;
ALTER TABLE books RENAME CONSTRAINT books_pkey TO book_pkey;
ALTER TABLE books RENAME TO book;
