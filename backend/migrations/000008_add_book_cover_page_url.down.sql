-- 先にページの制約を外す（書影だけ消すと「ページがあるのに Google Books の書影でない」行になり制約に反するため）
ALTER TABLE book DROP CONSTRAINT ck_book_cover_page_url;

-- Google Books の書影は、ページの URL が無いと規約上表示できないので外す（次に本を更新したときに取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'googlebooks';
ALTER TABLE book DROP COLUMN cover_page_url;

ALTER TABLE book DROP CONSTRAINT ck_book_cover_source;
ALTER TABLE book ADD CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd'));

COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd）。画面のクレジット表示に使う。cover_url と同時にNULL';
