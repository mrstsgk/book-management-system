-- CreateAuthorControllerTest用のデータベーステストデータ管理
-- このファイルは @Sql アノテーションで実行され、テストの前後でデータをクリーンアップします

-- 著者テーブルのデータをクリーンアップ
-- author_book テーブルとの外部キー制約がある場合は、そちらから先にクリーンアップ
DELETE FROM author_book WHERE author_id IS NOT NULL;
DELETE FROM author;

-- オートインクリメントのシーケンスをリセット
-- これにより、テスト間でIDが一貫して予測可能になります
ALTER SEQUENCE author_id_seq  RESTART WITH 1;
