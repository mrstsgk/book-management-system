-- 著者と書籍の中間テーブル（多対多対応）
CREATE TABLE author_book (
    author_id INTEGER NOT NULL REFERENCES author (id),
    book_id   INTEGER NOT NULL REFERENCES book(id),
    version   INTEGER NOT NULL,
    PRIMARY KEY (book_id, author_id)
);

-- テーブルとカラムにコメントを追加
COMMENT ON TABLE author_book IS '著者と書籍の中間テーブル（多対多対応）';
COMMENT ON COLUMN author_book.author_id IS '著者ID（外部キー）';
COMMENT ON COLUMN author_book.book_id IS '書籍ID（外部キー）';
COMMENT ON COLUMN author_book.version IS 'バージョン（楽観的ロック用）';
