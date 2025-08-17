-- CreateAuthorControllerTest用のデータベーステストデータ管理
-- このファイルは @Sql アノテーションで実行され、テストの前後でデータをクリーンアップします

-- 著者テーブルのデータをクリーンアップ
-- book_author テーブルとの外部キー制約がある場合は、そちらから先にクリーンアップ
DELETE FROM book_author WHERE author_id IS NOT NULL;
DELETE FROM author;

-- オートインクリメントのシーケンスをリセット
-- これにより、テスト間でIDが一貫して予測可能になります
ALTER SEQUENCE author_id_seq RESTART WITH 1;

-- テスト用の初期データ（必要に応じて追加）
-- このサンプルでは、既存データとの競合を避けるため、特に初期データは挿入しません
-- 必要に応じてテストケース固有のデータを挿入してください

-- 例: 特定のテストで既存データが必要な場合
-- INSERT INTO author (name, birth_date) VALUES ('テスト著者', '1900-01-01');