package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// sampleBook は見本データ1冊分。書誌は外部カタログに頼らず自前で持つ（ネットに繋がらなくても起動できるように）。
// 書影は起動時に外部カタログから取る（提供元の URL をそのまま表示する決まりのため、自前では持たない）。
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

// publisherLambdaNote は見本データに複数回出てくる出版社名（goconst 対策）。
const publisherLambdaNote = "ラムダノート"

// sampleBooks は実際に読んだ10冊。書誌は openBD の値を写し、openBD が誤っている著者・出版社は出版社の書誌ページで直した
// （データ指向アプリケーションデザイン第2版は openBD にも出版社のページにも無く、出版社・発売日を確認できなかったため空）。
// 一言まとめ・感想・評価は仮の値で、本人が書いた文章に差し替えるまで要件定義 §4.3「架空の感想は入れない」を満たさない。
var sampleBooks = []sampleBook{
	{
		isbn: "9784295016090", title: "AWS認定ソリューションアーキテクト-アソシエイト教科書 : 試験番号SAA-C03",
		authors: "鳥谷部昭寛, 宮口光平, 半田大樹, 株式会社ソキウス・ジャパン", publisher: "インプレス", publishedOn: "202303",
		titleOverride: "徹底攻略 AWS認定 ソリューションアーキテクト − アソシエイト教科書 第3版［SAA-C03］対応",
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
		authors: "戸根勤, 日経NETWORK（監修）", publisher: "日経BP", publishedOn: "200704",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784908686153", title: "事業をエンジニアリングする技術者たち フルサイクル開発者がつくるCARTAの現場",
		authors: "株式会社CARTA HOLDINGS 監修, 和田卓人 編集", publisher: publisherLambdaNote, publishedOn: "20220808",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784908686122", title: "Goならわかるシステムプログラミング 第2版",
		authors: "渋川よしき", publisher: publisherLambdaNote, publishedOn: "20220323",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784910313009", title: "Prometheus実践ガイド",
		authors: "仲亀拓馬", publisher: "テッキーメディア", publishedOn: "20220601",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784908686108", title: "Webブラウザセキュリティ Webアプリケーションの安全性を支える仕組みを整理する",
		authors: "米内貴志", publisher: publisherLambdaNote, publishedOn: "20210115",
		summary: placeholderText, comment: placeholderText, rating: 3,
	},
	{
		isbn: "9784908686207", title: "型システムのしくみ TypeScriptで実装しながら学ぶ型とプログラミング言語",
		authors: "遠藤侑介", publisher: publisherLambdaNote, publishedOn: "20250418",
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

// seedIfEmpty は本が1冊も無いときだけ見本データを保存する。
// 登録ユースケースは書誌を外部カタログから取る前提なので使わず、Repository に直接書く。
func seedIfEmpty(ctx context.Context, books domainbook.Repository, query domainbook.Query, catalog domainbook.BookCatalog) error {
	r, err := common.NewListRange(1, 0)
	if err != nil {
		return err
	}
	list, err := query.FindList(ctx, domainbook.ListCondition{}, r)
	if err != nil {
		return err
	}
	if list.Total > 0 {
		return nil
	}
	// 途中まで保存された状態を残さないよう、全冊を先に検証する
	prepared := make([]*domainbook.Book, 0, len(sampleBooks))
	for _, s := range sampleBooks {
		b, err := s.toBook()
		if err != nil {
			return fmt.Errorf("sample book %s: %w", s.isbn, err)
		}
		prepared = append(prepared, b)
	}
	for _, b := range prepared {
		b.Cover = lookupSampleCover(ctx, catalog, b.ISBN)
	}
	// 1冊ずつ保存すると、途中で失敗したとき次の起動では「本がある」と判断されて残りが入らないため、まとめて保存する
	if err := books.CreateAll(ctx, prepared); err != nil {
		if errors.Is(err, common.ErrConflict) {
			// 同時に起動した別のプロセスが先に入れた
			slog.InfoContext(ctx, "sample books already seeded by another process; skipping")
			return nil
		}
		return err
	}
	slog.InfoContext(ctx, "seeded sample books", "count", len(prepared))
	return nil
}

// lookupSampleCover は書影だけ外部カタログから取る。取れなければ書影なしにする（ネットが無くても起動できるように）。
func lookupSampleCover(ctx context.Context, catalog domainbook.BookCatalog, isbn domainbook.ISBN) *domainbook.Cover {
	entry, err := catalog.Lookup(ctx, isbn)
	if err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			slog.WarnContext(ctx, "sample book cover lookup failed; seeding without a cover", "isbn", isbn.String(), "error", err)
		}
		return nil
	}
	return entry.Cover
}
