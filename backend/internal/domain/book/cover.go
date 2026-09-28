package book

import (
	"fmt"
	"net/url"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// CoverSource は書影の提供元。画面に出すクレジット（楽天ウェブサービスは表記が必須）と、
// 取り直すときにどのサービスに問い合わせるかを決める。
type CoverSource string

const (
	CoverSourceOpenBD  CoverSource = "openbd"
	CoverSourceRakuten CoverSource = "rakuten"
)

const coverURLMaxLength = 2048

const (
	// RakutenRetention は楽天由来の情報を持てる期間。規約の「最長3か月」はどの月の組み合わせでも90日以上なので、
	// 90日で切れば必ず規約内に収まる。
	RakutenRetention = 90 * 24 * time.Hour
	// RakutenRefreshAfter を過ぎたら取り直す（期限の7日前）。
	RakutenRefreshAfter = 83 * 24 * time.Hour
)

// RakutenExpired は取得日時 fetchedAt の楽天由来の情報が now の時点で保持期限を過ぎているかを返す。
// 参照側（Read Model を組み立てる Query）も VO と同じ規則で判定するため、関数として公開する。
func RakutenExpired(fetchedAt, now time.Time) bool {
	return !now.Before(fetchedAt.Add(RakutenRetention))
}

// Cover は提供元がホストする書影の VO。提供元が認めているのは本の紹介のためにそのまま表示することだけなので、
// 画像そのものは保存・加工しない。楽天の書影は、規約上一緒に必要な商品ページと、保持期限の起点になる取得日時を必ず持つ。
type Cover struct {
	url        string
	source     CoverSource
	productURL string
	fetchedAt  time.Time
}

// NewCover は openBD の書影を作る。楽天の書影は商品ページと取得日時が要るので NewRakutenCover で作る。
func NewCover(rawURL string, source CoverSource) (Cover, error) {
	if source != CoverSourceOpenBD && source != CoverSourceRakuten {
		return Cover{}, fmt.Errorf("%w: 書影の提供元が不正です", common.ErrInvalid)
	}
	if !validCoverURL(rawURL) {
		return Cover{}, fmt.Errorf("%w: 書影のURLが不正です", common.ErrInvalid)
	}
	return Cover{url: rawURL, source: source}, nil
}

// NewRakutenCover は楽天の書影を、画像・商品ページ・取得日時が揃ったときだけ作る。
func NewRakutenCover(imageURL, productURL string, fetchedAt time.Time) (Cover, error) {
	if !validCoverURL(imageURL) {
		return Cover{}, fmt.Errorf("%w: 書影のURLが不正です", common.ErrInvalid)
	}
	if !validCoverURL(productURL) {
		return Cover{}, fmt.Errorf("%w: 楽天の商品ページのURLが不正です", common.ErrInvalid)
	}
	if fetchedAt.IsZero() {
		return Cover{}, fmt.Errorf("%w: 楽天の書影の取得日時がありません", common.ErrInvalid)
	}
	return Cover{url: imageURL, source: CoverSourceRakuten, productURL: productURL, fetchedAt: fetchedAt}, nil
}

func validCoverURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && len(raw) <= coverURLMaxLength && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}

func (c Cover) URL() string {
	return c.url
}

func (c Cover) Source() CoverSource {
	return c.source
}

// ProductURL は楽天の商品ページ。openBD の書影なら空。
func (c Cover) ProductURL() string {
	return c.productURL
}

// FetchedAt は楽天の書影を取得した日時。openBD の書影ならゼロ値。
func (c Cover) FetchedAt() time.Time {
	return c.fetchedAt
}

// IsExpired は楽天の書影が保持期限を過ぎているかを返す。openBD には保持期限が無いので常に false。
func (c Cover) IsExpired(now time.Time) bool {
	return c.source == CoverSourceRakuten && RakutenExpired(c.fetchedAt, now)
}

// NeedsRefresh は楽天の書影を取り直す時期かを返す。openBD は取り直す必要が無いので常に false。
func (c Cover) NeedsRefresh(now time.Time) bool {
	return c.source == CoverSourceRakuten && !now.Before(c.fetchedAt.Add(RakutenRefreshAfter))
}
