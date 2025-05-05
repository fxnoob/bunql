package pagination

import (
	"github.com/fxnoob/bunql/dto"
	"github.com/uptrace/bun"
)

// ApplyPagination applies pagination to the query
func ApplyPagination(query *bun.SelectQuery, p *dto.Pagination) *bun.SelectQuery {
	if p.Limit > 0 {
		query = query.Limit(p.Limit)

		// Calculate offset
		if p.Page > 0 {
			offset := (p.Page - 1) * p.Limit
			query = query.Offset(offset)
		}
	}

	return query
}
