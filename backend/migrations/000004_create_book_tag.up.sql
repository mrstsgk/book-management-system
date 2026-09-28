CREATE TABLE book_tag (
    book_id BIGINT NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, tag_id)
);

COMMENT ON TABLE book_tag IS '本と分野タグの多対多の中間テーブル。本・タグどちらが消えても自動で外れる';
COMMENT ON COLUMN book_tag.book_id IS '読んだ本のID';
COMMENT ON COLUMN book_tag.tag_id IS 'タグのID';
