-- UpdateBookControllerTest用のデータベーステストデータ管理
-- このファイルは @IntegrationTestWithSql アノテーションで実行され、テストの前後でデータをクリーンアップします

-- 書籍と著者の関連テーブルをクリーンアップ
TRUNCATE TABLE book_author CASCADE;

-- 書籍テーブルのデータをクリーンアップ
TRUNCATE TABLE book CASCADE;

-- 著者テーブルのデータをクリーンアップ  
TRUNCATE TABLE author CASCADE;

-- テスト用の著者データを挿入
INSERT INTO author (id, name, birth_date) VALUES
(1, '夏目漱石', '1867-02-09'),
(2, '太宰治', '1909-06-19'),
(3, '芥川龍之介', '1892-03-01');

-- テスト用の書籍データを挿入
INSERT INTO book (id, title, price, publish_status) VALUES
(1, '吾輩は猫である', 1500.00, 2),
(2, '日本文学選集', 3000.00, 1);

-- 書籍と著者の関連データを挿入
-- 書籍1: 夏目漱石のみ
INSERT INTO book_author (book_id, author_id) VALUES (1, 1);

-- 書籍2: 夏目漱石、太宰治、芥川龍之介（複数著者）
INSERT INTO book_author (book_id, author_id) VALUES 
(2, 1),
(2, 2),
(2, 3);

-- シーケンスをリセット
ALTER SEQUENCE author_id_seq RESTART WITH 4;
ALTER SEQUENCE book_id_seq RESTART WITH 3;