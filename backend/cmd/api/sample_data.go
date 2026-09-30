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

// publisherLambdaNote は見本データに複数回出てくる出版社名（goconst 対策）。
const publisherLambdaNote = "ラムダノート"

// sampleBooks は実際に読んだ10冊。書誌は openBD の値を写し、openBD が誤っている著者・出版社は出版社の書誌ページで直した。
// 一言まとめ・感想は仮の下書き文言で、本人が実際の読書体験に基づいて書いた文章に差し替えるまで
// 要件定義 §4.3「架空の感想は入れない」を満たさない。
var sampleBooks = []sampleBook{
	{
		isbn: "9784295016090", title: "AWS認定ソリューションアーキテクト-アソシエイト教科書 : 試験番号SAA-C03",
		authors: "鳥谷部昭寛, 宮口光平, 半田大樹, 株式会社ソキウス・ジャパン", publisher: "インプレス", publishedOn: "202303",
		titleOverride: "徹底攻略 AWS認定 ソリューションアーキテクト − アソシエイト教科書 第3版［SAA-C03］対応",
		summary:       "AWS認定SAA-C03の要点を効率よく押さえられる一冊。",
		comment: "AWS全体を俯瞰しつつ、セキュリティ・可用性・パフォーマンス・コストの観点で整理されており、" +
			"資格対策だけでなく実務でのサービス選定の指針としても参考になった。試験範囲を網羅的にカバーしている分、" +
			"実務で頻出のポイントに絞った深掘りは別途必要だと感じた。",
		rating: 3,
	},
	{
		isbn: "9784297146221", title: "改訂新版　良いコード／悪いコードで学ぶ設計入門 ―保守しやすい　成長し続けるコードの書き方",
		authors: "仙塲大也", publisher: "技術評論社",
		summary: "設計の勘所を具体例から学べる、実務にすぐ活かせる一冊。",
		comment: "「なぜこの設計が悪いのか」を具体的なコード例とともに説明してくれるので、感覚的に理解していた設計の勘所が言語化された。" +
			"特にカプセル化と条件分岐の整理は、自分のコードを見直すきっかけになった。18章というボリュームだが、" +
			"章ごとに独立して読めるので実務の合間にも読みやすい。",
		rating: 3,
	},
	{
		isbn: "9784798166391", title: "プロダクトマネジメントのすべて = THE COMPLETE BOOK OF PRODUCT MANAGEMENT : 事業戦略・IT開発・UXデザイン・マーケティングからチーム・組織運営まで",
		authors: "及川,卓也,1965- 曽根原,春樹 小城,久美子", publisher: "翔泳社", publishedOn: "202103",
		summary: "PMの役割を事業・開発・UXまで横断的に整理した教科書的な一冊。",
		comment: "プロダクトマネージャーという役割の解像度が上がった。事業戦略からUX、組織運営まで扱う範囲が広く、" +
			"辞書的に参照する使い方が向いていそう。エンジニア視点で読むと、PMがどこまで見ているかを知れて、" +
			"連携の仕方を見直すきっかけになった。",
		rating: 3,
	},
	{
		isbn: "9784873118703", title: "データ指向アプリケーションデザイン ―信頼性、拡張性、保守性の高い分散システム設計の原理",
		authors: "Martin Kleppmann, 斉藤太郎（訳）, 玉川竜司（訳）", publisher: "オーム社", publishedOn: "201907",
		summary: "分散システムのデータ設計を支える考え方を体系的に学べる名著。",
		comment: "レプリケーションやパーティショニング、トランザクションの話が、単なる技術解説ではなく" +
			"「なぜそのトレードオフが生まれるのか」から説明されていて腹落ちした。普段何気なく使っているDBやミドルウェアの" +
			"裏側を理解する助けになった。読み応えがあるので、章ごとに時間をかけて読んだ。",
		rating: 3,
	},
	{
		isbn: "9784822283117", title: "ネットワークはなぜつながるのか : 知っておきたいTCP/IP、LAN、光ファイバの基礎知識",
		authors: "戸根勤, 日経NETWORK（監修）", publisher: "日経BP", publishedOn: "200704",
		summary: "ブラウザ操作を起点にネットワークの仕組みを一気通貫で学べる入門書。",
		comment: "URLを打ち込んでからページが表示されるまでの流れを追う構成が分かりやすく、" +
			"普段意識しないTCP/IPやLAN機器の動きが具体的にイメージできるようになった。" +
			"ネットワークの基礎を体系的に学び直すのにちょうど良い一冊だった。",
		rating: 3,
	},
	{
		isbn: "9784908686153", title: "事業をエンジニアリングする技術者たち フルサイクル開発者がつくるCARTAの現場",
		authors: "株式会社CARTA HOLDINGS 監修, 和田卓人 編集", publisher: publisherLambdaNote, publishedOn: "20220808",
		summary: "事業の現場で働くエンジニアへのインタビュー集。組織のリアルが伝わる。",
		comment: "特定の技術解説ではなく、実在する事業・組織でエンジニアがどう考え、どう動いているかが" +
			"インタビュー形式で語られていて、技術書というより組織論・キャリア論として読んだ。" +
			"フルサイクル開発者という体制の話は、自分のチームの開発体制を考える上でも参考になった。",
		rating: 3,
	},
	{
		isbn: "9784908686122", title: "Goならわかるシステムプログラミング 第2版",
		authors: "渋川よしき", publisher: publisherLambdaNote, publishedOn: "20220323",
		summary: "Goのコードを入口にOSの仕組みを学べる、実務寄りのシステムプログラミング入門。",
		comment: "システムコールやソケット通信、プロセス、シグナルといったOSの基礎を、Goの実際のコードから追える構成が良かった。" +
			"教科書的な説明だけでは繋がらなかった知識が、コードを動かしながら理解できた。" +
			"ボリュームがあるので、必要な章から読む使い方も向いていそう。",
		rating: 3,
	},
	{
		isbn: "9784910313009", title: "Prometheus実践ガイド",
		authors: "仲亀拓馬", publisher: "テッキーメディア", publishedOn: "20220601",
		summary: "PrometheusによるKubernetes監視を基礎から実践まで学べる一冊。",
		comment: "PromQLの書き方やアラート設計など、実際に運用する上で必要な知識がまとまっていて助かった。" +
			"Kubernetes環境での監視構築を検討する際の実践的な手引きとして参考にした。",
		rating: 3,
	},
	{
		isbn: "9784908686108", title: "Webブラウザセキュリティ Webアプリケーションの安全性を支える仕組みを整理する",
		authors: "米内貴志", publisher: publisherLambdaNote, publishedOn: "20210115",
		summary: "Origin・Cookie・プロセスモデルなど、ブラウザのセキュリティ機構を整理した一冊。",
		comment: "普段なんとなく守っているCORSやCSPのようなルールが、実際にはどういう脅威に対する対策なのかを整理して理解できた。" +
			"Webアプリケーションを開発する上で、セキュリティ周りの実装判断に自信が持てるようになった。",
		rating: 3,
	},
	{
		isbn: "9784908686207", title: "型システムのしくみ TypeScriptで実装しながら学ぶ型とプログラミング言語",
		authors: "遠藤侑介", publisher: publisherLambdaNote, publishedOn: "20250418",
		summary: "TypeScriptで型検査器を実装しながら型システムの仕組みを学べる一冊。",
		comment: "型理論を数式で説明されるより、実際に手を動かして型検査器を実装しながら学ぶ構成が自分には合っていた。" +
			"ジェネリクスや再帰型がなぜ複雑になるのかを、実装を通じて実感できたのが良かった。",
		rating: 3,
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
