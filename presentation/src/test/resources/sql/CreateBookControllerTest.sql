-- CreateBookControllerTest用のデータベーステストデータ管理

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

-- シーケンスをリセット
ALTER SEQUENCE author_id_seq RESTART WITH 4;
ALTER SEQUENCE book_id_seq RESTART WITH 1;