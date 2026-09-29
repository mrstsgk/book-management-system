ALTER TABLE book DROP CONSTRAINT IF EXISTS ck_book_rakuten_cover;
ALTER TABLE book DROP COLUMN IF EXISTS cover_fetched_at;
ALTER TABLE book DROP COLUMN IF EXISTS cover_product_url;
