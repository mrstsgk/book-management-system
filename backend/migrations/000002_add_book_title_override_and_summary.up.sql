ALTER TABLE book ADD COLUMN title_override VARCHAR(255);
-- 既存の行があっても足せるよう既定値付きで足し、すぐ外す（以後は必ずアプリが値を入れる）
ALTER TABLE book ADD COLUMN summary VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE book ALTER COLUMN summary DROP DEFAULT;

COMMENT ON COLUMN book.title_override IS '自分で上書きした書名（1〜255文字）。NULLなら上書きしていない。外部カタログを取り直しても消えない';
COMMENT ON COLUMN book.summary IS '一覧で見せる一言まとめ（1〜100文字、改行なし）';
COMMENT ON COLUMN book.title IS '書名（外部カタログの値）。画面では title_override があればそちらを見せる';
