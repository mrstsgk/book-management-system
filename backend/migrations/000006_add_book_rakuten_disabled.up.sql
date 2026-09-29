ALTER TABLE book ADD COLUMN rakuten_disabled BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN book.rakuten_disabled IS '楽天から削除の指示を受けた本なら TRUE。以後この本には楽天の書影を付けない';
