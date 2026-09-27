ALTER TABLE book
    DROP COLUMN IF EXISTS image_key,
    DROP COLUMN IF EXISTS amazon_url;
