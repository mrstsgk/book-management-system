package main

import (
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// sampleBook は見本データ1冊分。書誌は外部カタログに頼らず自前で持つ（ネットに繋がらなくても起動できるように）。
// 楽天由来の情報は保持期限を守れないため持たない（書影は起動時に外部カタログから取る）。
type sampleBook struct {
	isbn        string
	title       string
	authors     string
	publisher   string
	publishedOn string
	// titleOverride は外部カタログの書名が実際と違うときの上書き。空なら上書きしない。
	titleOverride string
	summary       string
	comment       string
	rating        int
}

// placeholderText は本人が一言まとめ・感想を書くまでの仮の文言。
const placeholderText = "（準備中）"

// sampleBooks は実際に読んだ5冊。書誌は openBD の値を写したもの（データ指向アプリケーションデザイン第2版は
// openBD に無く、出版社・発売日を確認できなかったため空）。
// 一言まとめ・感想・評価は仮の値で、本人が書いた文章に差し替えるまで要件定義 §4.3「架空の感想は入れない」を満たさない。
var sampleBooks = []sampleBook{
	{
		isbn: "9784295016090", title: "AWS認定ソリューションアーキテクト-アソシエイト教科書 : 試験番号SAA-C03",
		publisher: "インプレス", publishedOn: "202303",
		titleOverride: "徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版",
		summary:       placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784297146221", title: "改訂新版　良いコード／悪いコードで学ぶ設計入門 ―保守しやすい　成長し続けるコードの書き方",
		authors: "仙塲大也", publisher: "技術評論社",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784798166391", title: "プロダクトマネジメントのすべて = THE COMPLETE BOOK OF PRODUCT MANAGEMENT : 事業戦略・IT開発・UXデザイン・マーケティングからチーム・組織運営まで",
		authors: "及川,卓也,1965- 曽根原,春樹 小城,久美子", publisher: "翔泳社", publishedOn: "202103",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784814401802", title: "データ指向アプリケーションデザイン 第2版 ―信頼性、拡張性、保守性の高い分散システム設計の原理",
		authors: "Martin Kleppmann",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784822283117", title: "ネットワークはなぜつながるのか : 知っておきたいTCP/IP、LAN、光ファイバの基礎知識",
		authors: "戸根,勤 日経BP", publisher: "日経BP出版センター", publishedOn: "200704",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
}

// toBook は見本データを VO で検証して、保存前の本にする。
func (s sampleBook) toBook() (*domainbook.Book, error) {
	isbn, err := domainbook.NewISBN(s.isbn)
	if err != nil {
		return nil, err
	}
	bib, err := domainbook.NewBibliography(s.title, s.authors, s.publisher, s.publishedOn)
	if err != nil {
		return nil, err
	}
	summary, err := domainbook.NewSummary(s.summary)
	if err != nil {
		return nil, err
	}
	comment, err := domainbook.NewComment(s.comment)
	if err != nil {
		return nil, err
	}
	rating, err := domainbook.NewRating(s.rating)
	if err != nil {
		return nil, err
	}
	b := domainbook.New(isbn, bib, nil, summary, comment, rating, domainbook.TagSelection{})
	if s.titleOverride != "" {
		title, err := domainbook.NewTitle(s.titleOverride)
		if err != nil {
			return nil, err
		}
		b.OverrideTitle(&title)
	}
	return b, nil
}
