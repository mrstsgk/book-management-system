-- 楽天ブックスは撤去した（書影は Google Books → openBD の順で取る。docs/adr/2026-09-29-cover-from-google-books-and-openbd.md）。
-- 楽天の書影は規約上そのまま持ち続けられないので外す（次に本を更新したときに openBD から取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'rakuten';

ALTER TABLE book DROP CONSTRAINT ck_book_rakuten_cover;
ALTER TABLE book DROP COLUMN cover_product_url;
ALTER TABLE book DROP COLUMN cover_fetched_at;
ALTER TABLE book DROP COLUMN rakuten_disabled;

ALTER TABLE book DROP CONSTRAINT ck_book_cover_source;
ALTER TABLE book ADD CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd'));

COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd）。画面のクレジット表示に使う。cover_url と同時にNULL';
