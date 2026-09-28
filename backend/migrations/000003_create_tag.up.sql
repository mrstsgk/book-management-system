CREATE TABLE tag (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(30) NOT NULL,
    version    INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tag_name UNIQUE (name)
);

COMMENT ON TABLE tag IS '自分が定義した分野タグ。本には book_tag 経由で複数付けられる';
COMMENT ON COLUMN tag.id IS 'タグのID（主キー）';
COMMENT ON COLUMN tag.name IS 'タグ名（1〜30文字）。一意';
COMMENT ON COLUMN tag.version IS 'バージョン（楽観的ロック用）。作成時1、更新ごとに+1';
COMMENT ON COLUMN tag.created_at IS '作成日時';
COMMENT ON COLUMN tag.updated_at IS '最終更新日時';
