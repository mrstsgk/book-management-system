ALTER TABLE book ADD COLUMN cover_product_url VARCHAR(2048);
ALTER TABLE book ADD COLUMN cover_fetched_at TIMESTAMPTZ;

-- 既存の楽天の書影は商品ページも取得日時も分からないので外す（本を更新すれば取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'rakuten';

ALTER TABLE book ADD CONSTRAINT ck_book_rakuten_cover CHECK (
    (cover_source IS NOT DISTINCT FROM 'rakuten') = (cover_product_url IS NOT NULL AND cover_fetched_at IS NOT NULL)
);

COMMENT ON COLUMN book.cover_product_url IS '楽天の商品ページの URL（楽天の書影のときだけ。楽天の規約上、書影と一緒にリンクする）。それ以外は NULL';
COMMENT ON COLUMN book.cover_fetched_at IS '楽天の書影・商品ページを取得した日時（保持期限90日の起点）。楽天の書影のときだけ。それ以外は NULL';
