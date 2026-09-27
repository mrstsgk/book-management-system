package author

import (
	"context"

	"gorm.io/gorm"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
)

type query struct {
	db *gorm.DB
}

func NewQuery(db *gorm.DB) domainauthor.Query {
	return &query{db: db}
}

func (q *query) Exists(ctx context.Context, id domainauthor.ID) (bool, error) {
	var n int64
	if err := q.db.WithContext(ctx).Model(&model{}).Where("id = ?", int64(id)).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
