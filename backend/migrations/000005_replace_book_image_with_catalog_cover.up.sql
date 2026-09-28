-- 書影は自前でアップロードせず、外部カタログ（openBD・楽天ブックス）の URL を表示する方式に変えるため、
-- 画像キーと手入力の Amazon リンクを削除し、ISBN と書影の URL・提供元を持つ。
-- Amazon のリンクは ISBN から導出するので保存しない。
ALTER TABLE book
    DROP COLUMN IF EXISTS image_key,
    DROP COLUMN IF EXISTS amazon_url,
    ADD COLUMN isbn         VARCHAR(13) UNIQUE,
    ADD COLUMN cover_url    VARCHAR(2048),
    ADD COLUMN cover_source VARCHAR(16),
    ADD CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd', 'rakuten')),
    ADD CONSTRAINT ck_book_cover_pair CHECK ((cover_url IS NULL) = (cover_source IS NULL));

COMMENT ON COLUMN book.isbn IS 'ISBN（13桁。ISBN-10 は 978 付きの13桁に変換して保存）。未設定はNULL。設定された値は一意';
COMMENT ON COLUMN book.cover_url IS '書影の URL（提供元がホストする画像。自前では保存しない）。未取得はNULL';
COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd / rakuten）。画面のクレジット表示に使う。cover_url と同時にNULL';
