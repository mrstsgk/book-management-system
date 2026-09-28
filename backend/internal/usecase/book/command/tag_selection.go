package command

import (
	"context"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// parseTagSelection はタグIDの入力を検証し、実在するかも確かめる。
func parseTagSelection(ctx context.Context, tags tag.Query, rawIDs []int64) (book.TagSelection, error) {
	ids := make([]tag.ID, len(rawIDs))
	for i, v := range rawIDs {
		ids[i] = tag.ID(v)
	}
	selection, err := book.NewTagSelection(ids)
	if err != nil {
		return book.TagSelection{}, err
	}
	ok, err := tags.ExistsAll(ctx, ids)
	if err != nil {
		return book.TagSelection{}, err
	}
	if !ok {
		return book.TagSelection{}, fmt.Errorf("%w: 存在しない分野タグが含まれています", common.ErrInvalid)
	}
	return selection, nil
}
