-- 書籍テーブル
CREATE TABLE IF NOT EXISTS book (
    id             SERIAL PRIMARY KEY,
    title          VARCHAR(255) NOT NULL,
    price          NUMERIC(10, 2) NOT NULL,
    publish_status INTEGER NOT NULL,
    version        INTEGER NOT NULL
);

-- テーブルとカラムにコメントを追加
COMMENT ON TABLE book IS '書籍情報を管理するテーブル';
COMMENT ON COLUMN book.id IS '書籍ID（主キー）';
COMMENT ON COLUMN book.title IS '書籍タイトル';
COMMENT ON COLUMN book.price IS '価格（円）';
COMMENT ON COLUMN book.publish_status IS '出版状況（1: 未出版, 2: 出版済み）';
COMMENT ON COLUMN book.version IS 'バージョン（楽観的ロック用）';

