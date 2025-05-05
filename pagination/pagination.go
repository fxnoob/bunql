package pagination

import (
	"github.com/fxnoob/bunql/dto"
	"github.com/uptrace/bun"
)

// ApplyPagination applies pagination to the query
func ApplyPagination(query *bun.SelectQuery, p *dto.Pagination) *bun.SelectQuery {
	if p.PageSize > 0 {
		query = query.Limit(p.PageSize)

		// Calculate offset
		if p.Page > 0 {
			offset := (p.Page - 1) * p.PageSize
			query = query.Offset(offset)
		}
	}

	return query
}
