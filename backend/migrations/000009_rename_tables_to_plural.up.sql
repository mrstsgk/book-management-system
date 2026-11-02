-- テーブル名を規則（業務の概念・複数形・チャネルの接頭辞を付けない）に揃える。
-- 制約・シーケンスの名前もテーブル名を含むものは揃える（主キー・一意制約の RENAME CONSTRAINT は裏の索引の名前も変える）。
ALTER TABLE book RENAME TO books;
ALTER TABLE books RENAME CONSTRAINT book_pkey TO books_pkey;
ALTER TABLE books RENAME CONSTRAINT uq_book_isbn TO uq_books_isbn;
ALTER TABLE books RENAME CONSTRAINT ck_book_rating TO ck_books_rating;
ALTER TABLE books RENAME CONSTRAINT ck_book_cover_source TO ck_books_cover_source;
ALTER TABLE books RENAME CONSTRAINT ck_book_cover_pair TO ck_books_cover_pair;
ALTER SEQUENCE book_id_seq RENAME TO books_id_seq;

ALTER TABLE tag RENAME TO tags;
ALTER TABLE tags RENAME CONSTRAINT tag_pkey TO tags_pkey;
ALTER TABLE tags RENAME CONSTRAINT uq_tag_name TO uq_tags_name;
ALTER SEQUENCE tag_id_seq RENAME TO tags_id_seq;

ALTER TABLE book_tag RENAME TO book_tags;
ALTER TABLE book_tags RENAME CONSTRAINT book_tag_pkey TO book_tags_pkey;
ALTER TABLE book_tags RENAME CONSTRAINT book_tag_book_id_fkey TO book_tags_book_id_fkey;
ALTER TABLE book_tags RENAME CONSTRAINT book_tag_tag_id_fkey TO book_tags_tag_id_fkey;

ALTER TABLE admin_session RENAME TO sessions;
ALTER TABLE sessions RENAME CONSTRAINT admin_session_pkey TO sessions_pkey;
