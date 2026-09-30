CREATE TABLE admin_session (
    id                  VARCHAR(64) PRIMARY KEY,
    expires_at          TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE admin_session IS '管理画面のログイン済みセッション。利用者は自分 1 人なので誰のセッションかは持たない。パスワードもハッシュも入れない';
COMMENT ON COLUMN admin_session.id IS 'セッションID（32 byte の乱数を base64url にした 43 文字）。ブラウザには httpOnly Cookie でこれだけを渡す';
COMMENT ON COLUMN admin_session.expires_at IS 'アイドル期限。書き込み API を通るたびに「今 + 1 時間」に延びる（絶対期限は超えない）';
COMMENT ON COLUMN admin_session.absolute_expires_at IS '絶対期限。ログイン時刻 + 24 時間で固定';
COMMENT ON COLUMN admin_session.created_at IS 'ログイン日時';
