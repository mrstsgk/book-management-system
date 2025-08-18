-- UpdateAuthorControllerTest用のデータベーステストデータ管理
-- このファイルは @IntegrationTestWithSql アノテーションで実行され、テストの前後でデータをクリーンアップします

-- 著者テーブルのデータをクリーンアップ
-- book_author テーブルとの外部キー制約がある場合は、そちらから先にクリーンアップ
DELETE FROM book_author WHERE author_id IS NOT NULL;
DELETE FROM author;

-- オートインクリメントのシーケンスをリセット
-- これにより、テスト間でIDが一貫して予測可能になります
ALTER SEQUENCE author_id_seq RESTART WITH 1;

-- 著者更新テスト用のテストデータを挿入
INSERT INTO author (id, name, birth_date, version) VALUES (1, '夏目漱石', '1867-02-09', 1);