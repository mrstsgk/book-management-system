-- 書籍と著者の中間テーブル（多対多対応）
CREATE TABLE book_author (
    book_id   INTEGER NOT NULL REFERENCES book(id),
    author_id INTEGER NOT NULL REFERENCES author (id),
    version   INTEGER NOT NULL,
    PRIMARY KEY (book_id, author_id)
);

-- テーブルとカラムにコメントを追加
COMMENT ON TABLE book_author IS '書籍と著者の中間テーブル（多対多対応）';
COMMENT ON COLUMN book_author.book_id IS '書籍ID（外部キー）';
COMMENT ON COLUMN book_author.author_id IS '著者ID（外部キー）';
COMMENT ON COLUMN book_author.version IS 'バージョン（楽観的ロック用）';
