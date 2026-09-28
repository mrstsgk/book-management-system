CREATE TABLE IF NOT EXISTS book (
    id           BIGSERIAL PRIMARY KEY,
    isbn         VARCHAR(13) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    authors      VARCHAR(500) NOT NULL DEFAULT '',
    publisher    VARCHAR(255) NOT NULL DEFAULT '',
    published_on VARCHAR(32) NOT NULL DEFAULT '',
    cover_url    VARCHAR(2048),
    cover_source VARCHAR(16),
    comment      TEXT NOT NULL,
    rating       SMALLINT NOT NULL,
    version      INTEGER NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_book_isbn UNIQUE (isbn),
    CONSTRAINT ck_book_rating CHECK (rating BETWEEN 1 AND 5),
    CONSTRAINT ck_book_cover_source CHECK (cover_source IN ('openbd', 'rakuten')),
    CONSTRAINT ck_book_cover_pair CHECK ((cover_url IS NULL) = (cover_source IS NULL))
);

COMMENT ON TABLE book IS '自分が読んだ本。書誌・書影は ISBN で外部カタログから取得し、感想と評価は自分で書く';
COMMENT ON COLUMN book.id IS '読んだ本のID（主キー）';
COMMENT ON COLUMN book.isbn IS 'ISBN（13桁。ISBN-10 は 978 付きの13桁に変換して保存）。一意';
COMMENT ON COLUMN book.title IS '書名（外部カタログの値）';
COMMENT ON COLUMN book.authors IS '著者（外部カタログの文字列のまま。訳者も混ざりうる）。無ければ空文字';
COMMENT ON COLUMN book.publisher IS '出版社（外部カタログの値）。無ければ空文字';
COMMENT ON COLUMN book.published_on IS '発売日（外部カタログの表記のまま。例: 201907）。無ければ空文字';
COMMENT ON COLUMN book.cover_url IS '書影の URL（提供元がホストする画像。自前では保存しない）。無ければNULL';
COMMENT ON COLUMN book.cover_source IS '書影の提供元（openbd / rakuten）。画面のクレジット表示に使う。cover_url と同時にNULL';
COMMENT ON COLUMN book.comment IS '自分が書いた感想（1〜5000文字）';
COMMENT ON COLUMN book.rating IS '自分が付けた評価（1〜5）';
COMMENT ON COLUMN book.version IS 'バージョン（楽観的ロック用）。作成時1、更新ごとに+1';
COMMENT ON COLUMN book.created_at IS '登録日時。一覧を新しく登録した順に並べるのに使う';
COMMENT ON COLUMN book.updated_at IS '最終更新日時';
