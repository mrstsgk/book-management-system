ALTER TABLE book
    ADD COLUMN amazon_url VARCHAR(2048),
    ADD COLUMN image_key  VARCHAR(255);

COMMENT ON COLUMN book.amazon_url IS 'Amazonの商品リンク（https・Amazonのホストのみ）。未設定はNULL';
COMMENT ON COLUMN book.image_key IS '表紙画像のオブジェクトストレージ上のキー。未設定はNULL';
