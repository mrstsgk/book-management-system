-- 著者テーブル
CREATE TABLE IF NOT EXISTS author (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    birth_date  DATE,
    version     INTEGER NOT NULL
);

-- テーブルとカラムにコメントを追加
COMMENT ON TABLE author IS '著者情報を管理するテーブル';
COMMENT ON COLUMN author.id IS '著者ID（主キー）';
COMMENT ON COLUMN author.name IS '著者名';
COMMENT ON COLUMN author.birth_date IS '生年月日';
COMMENT ON COLUMN author.version IS 'バージョン（楽観的ロック用）';

