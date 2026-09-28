COMMENT ON COLUMN book.title IS '書名（外部カタログの値）';
ALTER TABLE book DROP COLUMN IF EXISTS summary;
ALTER TABLE book DROP COLUMN IF EXISTS title_override;
