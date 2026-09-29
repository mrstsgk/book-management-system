ALTER TABLE book ADD COLUMN cover_product_url VARCHAR(2048);
ALTER TABLE book ADD COLUMN cover_fetched_at TIMESTAMPTZ;

-- 既存の楽天の書影は商品ページも取得日時も分からないので外す（本を更新すれば取り直される）
UPDATE book SET cover_url = NULL, cover_source = NULL WHERE cover_source = 'rakuten';

-- 楽天の書影なら商品ページ・取得日時の両方が必須、それ以外（openBD・書影なし）なら両方 NULL。
-- 等号で比べる書き方は NULL が混ざると CHECK が素通りする（NULL は違反扱いされない）ので、両側を明示する
ALTER TABLE book ADD CONSTRAINT ck_book_rakuten_cover CHECK (
    (cover_source = 'rakuten' AND cover_product_url IS NOT NULL AND cover_fetched_at IS NOT NULL)
    OR (cover_source IS DISTINCT FROM 'rakuten' AND cover_product_url IS NULL AND cover_fetched_at IS NULL)
);

COMMENT ON COLUMN book.cover_product_url IS '楽天の商品ページの URL（楽天の書影のときだけ。楽天の規約上、書影と一緒にリンクする）。それ以外は NULL';
COMMENT ON COLUMN book.cover_fetched_at IS '楽天の書影・商品ページを取得した日時（保持期限89日の起点）。楽天の書影のときだけ。それ以外は NULL';
