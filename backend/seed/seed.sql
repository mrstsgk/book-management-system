-- ローカル開発用の見本データ（`make seed` で投入。何度実行しても同じ状態になる）。
-- 画像キー seed/books/*.png は LocalStack の起動時に backend/localstack/seed/ からアップロードされる。
-- ID は 9001 番台に固定しており、同じ ID の既存行は見本データで上書きされる。
BEGIN;

INSERT INTO author (id, name, birth_date, version) VALUES
    (9001, '太宰治', '1909-06-19', 1),
    (9002, '芥川龍之介', '1892-03-01', 1),
    (9003, '夏目漱石', '1867-02-09', 1),
    (9004, '宮沢賢治', NULL, 1)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name, birth_date = EXCLUDED.birth_date, version = EXCLUDED.version;

INSERT INTO book (id, title, price, publish_status, amazon_url, image_key, version) VALUES
    (9001, '人間失格', 500, 2, 'https://www.amazon.co.jp/s?k=%E4%BA%BA%E9%96%93%E5%A4%B1%E6%A0%BC', 'seed/books/9001.png', 1),
    (9002, '走れメロス', 400, 2, NULL, 'seed/books/9002.png', 1),
    (9003, '羅生門', 450, 2, 'https://www.amazon.co.jp/s?k=%E7%BE%85%E7%94%9F%E9%96%80', NULL, 1),
    (9004, '吾輩は猫である', 800, 2, 'https://www.amazon.co.jp/s?k=%E5%90%BE%E8%BC%A9%E3%81%AF%E7%8C%AB%E3%81%A7%E3%81%82%E3%82%8B', 'seed/books/9004.png', 1),
    (9005, '銀河鉄道の夜', 600, 1, NULL, NULL, 1),
    (9006, '文豪短編集', 1200, 1, NULL, 'seed/books/9006.png', 1)
ON CONFLICT (id) DO UPDATE
    SET title = EXCLUDED.title, price = EXCLUDED.price, publish_status = EXCLUDED.publish_status,
        amazon_url = EXCLUDED.amazon_url, image_key = EXCLUDED.image_key, version = EXCLUDED.version;

-- 見本の書籍の著者の関連は毎回作り直す（手で付け替えた関連も見本の状態に戻す）。
DELETE FROM author_book WHERE book_id BETWEEN 9001 AND 9006;
INSERT INTO author_book (author_id, book_id, version) VALUES
    (9001, 9001, 1),
    (9001, 9002, 1),
    (9002, 9003, 1),
    (9003, 9004, 1),
    (9004, 9005, 1),
    (9001, 9006, 1),
    (9002, 9006, 1),
    (9003, 9006, 1);

-- ID を明示して挿入したため、以降の採番が見本データの ID と衝突しないよう進めておく。
DO $$
BEGIN
    PERFORM setval(pg_get_serial_sequence('author', 'id'), (SELECT MAX(id) FROM author));
    PERFORM setval(pg_get_serial_sequence('book', 'id'), (SELECT MAX(id) FROM book));
END $$;

COMMIT;
