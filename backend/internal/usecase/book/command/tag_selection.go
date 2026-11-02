package command

import (
	"context"
	"errors"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// parseTagSelection はタグIDの入力を検証し、実在するかも確かめる。違反は vs に集め、error は DB 障害だけ。
// 上限超過・重複のときは実在確認をしない（前提が成立しない検査は実行しない）。
func parseTagSelection(ctx context.Context, vs *common.Violations, tags tag.Query, rawIDs []int64) (book.TagSelection, error) {
	ids := make([]tag.ID, len(rawIDs))
	for i, v := range rawIDs {
		ids[i] = tag.ID(v)
	}
	selection, err := book.NewTagSelection(ids)
	if err != nil {
		rule := "invalid"
		switch {
		case errors.Is(err, book.ErrTooManyTags):
			rule = "max"
		case errors.Is(err, book.ErrDuplicateTags):
			rule = "tagIdsDuplicated"
		}
		vs.Add("tagIds", rule)
		return book.TagSelection{}, nil //nolint:nilerr // 違反は vs に集めて呼び出し側がまとめて返す。error は DB 障害だけ
	}
	ok, err := tags.ExistsAll(ctx, ids)
	if err != nil {
		return book.TagSelection{}, err
	}
	if !ok {
		vs.Add("tagIds", "tagIdsNotFound")
	}
	return selection, nil
}

// reviewFields は登録・更新に共通する感想まわりの VO。
type reviewFields struct {
	summary  book.Summary
	override *book.Title
	comment  book.Comment
	rating   book.Rating
}

// parseReviewFields は一言まとめ・書名の上書き・感想・評価を VO にする。違反は項目ごとに vs に集める
// （1 つ目で止めると、利用者が直しても次の項目でまた弾かれるため）。
func parseReviewFields(vs *common.Violations, summary, override, comment string, rating int) reviewFields {
	var f reviewFields
	var err error
	if f.summary, err = book.NewSummary(summary); err != nil {
		vs.Add("summary", "invalid")
	}
	if f.override, err = parseTitleOverride(override); err != nil {
		vs.Add("titleOverride", "invalid")
	}
	if f.comment, err = book.NewComment(comment); err != nil {
		vs.Add("comment", "invalid")
	}
	if f.rating, err = book.NewRating(rating); err != nil {
		vs.Add("rating", "invalid")
	}
	return f
}
