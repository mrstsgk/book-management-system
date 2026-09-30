-- 列と制約だけを戻す。up で外した楽天の書影は戻せない（取り直すには楽天から再取得が要る）
ALTER TABLE book DROP CONSTRAINT ck_book_cover_source;
ALTER TABLE book ADD CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd', 'rakuten'));
COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd / rakuten）。画面のクレジット表示に使う。cover_url と同時にNULL';

ALTER TABLE book ADD COLUMN rakuten_disabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE book ADD COLUMN cover_fetched_at TIMESTAMPTZ;
ALTER TABLE book ADD COLUMN cover_product_url VARCHAR(2048);
ALTER TABLE book ADD CONSTRAINT ck_book_rakuten_cover CHECK (
    (cover_source = 'rakuten' AND cover_product_url IS NOT NULL AND cover_fetched_at IS NOT NULL)
    OR (cover_source IS DISTINCT FROM 'rakuten' AND cover_product_url IS NULL AND cover_fetched_at IS NULL)
);

COMMENT ON COLUMN book.cover_product_url IS '楽天の商品ページの URL（楽天の書影のときだけ。楽天の規約上、書影と一緒にリンクする）。それ以外は NULL';
COMMENT ON COLUMN book.cover_fetched_at IS '楽天の書影・商品ページを取得した日時（保持期限89日の起点）。楽天の書影のときだけ。それ以外は NULL';
COMMENT ON COLUMN book.rakuten_disabled IS '楽天から削除の指示を受けた本なら TRUE。以後この本には楽天の書影を付けない';
