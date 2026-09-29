-- 書影を Google Books → openBD の順で取る（docs/adr/2026-09-29-cover-from-google-books-and-openbd.md）。
-- Google Books は規約上、書影と一緒にその本の Google Books のページへのリンクが要るので、ページの URL を持つ
ALTER TABLE book ADD COLUMN cover_page_url VARCHAR(2048);

ALTER TABLE book DROP CONSTRAINT ck_book_cover_source;
ALTER TABLE book ADD CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd', 'googlebooks'));
-- 「cover_source = 'googlebooks'」だと書影なし（NULL）の行で NULL になり CHECK を素通りするので、両辺が必ず真偽になる形で書く
ALTER TABLE book ADD CONSTRAINT ck_book_cover_page_url CHECK (
    (cover_page_url IS NOT NULL) = (cover_source IS NOT DISTINCT FROM 'googlebooks')
);

COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd / googlebooks）。画面のクレジット表示に使う。cover_url と同時にNULL';
COMMENT ON COLUMN book.cover_page_url IS '書影の提供元にあるその本のページの URL（Google Books の書影のときだけ。規約上、書影と一緒にリンクする）。それ以外は NULL';
