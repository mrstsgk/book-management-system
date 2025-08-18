-- UpdateBookControllerTest用のデータベーステストデータ管理
-- このファイルは @IntegrationTestWithSql アノテーションで実行され、テストの前後でデータをクリーンアップします

-- 書籍と著者の関連テーブルをクリーンアップ
TRUNCATE TABLE book_author CASCADE;

-- 書籍テーブルのデータをクリーンアップ
TRUNCATE TABLE book CASCADE;

-- 著者テーブルのデータをクリーンアップ  
TRUNCATE TABLE author CASCADE;

-- テスト用の著者データを挿入
INSERT INTO author (id, name, birth_date, version) VALUES
(1, '夏目漱石', '1867-02-09', 1),
(2, '太宰治', '1909-06-19', 1),
(3, '芥川龍之介', '1892-03-01', 1);

-- テスト用の書籍データを挿入
INSERT INTO book (id, title, price, publish_status, version) VALUES
(1, '吾輩は猫である', 1500.00, 2, 1),
(2, '日本文学選集', 3000.00, 1, 1);

-- 書籍と著者の関連データを挿入
-- 書籍1: 夏目漱石のみ
INSERT INTO book_author (book_id, author_id, version) VALUES (1, 1, 1);

-- 書籍2: 夏目漱石、太宰治、芥川龍之介（複数著者）
INSERT INTO book_author (book_id, author_id, version) VALUES 
(2, 1, 1),
(2, 2, 1),
(2, 3, 1);

-- シーケンスをリセット
ALTER SEQUENCE author_id_seq RESTART WITH 4;
ALTER SEQUENCE book_id_seq RESTART WITH 3;