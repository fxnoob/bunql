package bunql

import (
	"context"
	"github.com/fxnoob/bunql/dto"
	"github.com/fxnoob/bunql/filter"
	"github.com/fxnoob/bunql/pagination"
	"github.com/fxnoob/bunql/sorting"
	"github.com/uptrace/bun"
)

type BunQL struct {
	Filters    dto.FilterGroup
	Sort       []dto.SortField
	Pagination *dto.Pagination
}

// New creates a new BunQL instance
func New() *BunQL {
	return &BunQL{
		Filters: dto.FilterGroup{
			Logic:   "and",
			Filters: []dto.Filter{},
			Groups:  []dto.FilterGroup{},
		},
		Sort:       []dto.SortField{},
		Pagination: nil,
	}
}

// WithFilters adds filter to the query
func (q *BunQL) WithFilters(filters dto.FilterGroup) *BunQL {
	q.Filters = filters
	return q
}

// WithSort adds sorting to the query
func (q *BunQL) WithSort(sort []dto.SortField) *BunQL {
	q.Sort = sort
	return q
}

// WithPagination adds pagination to the query
func (q *BunQL) WithPagination(pagination *dto.Pagination) *BunQL {
	q.Pagination = pagination
	return q
}

// Apply applies all filter, sorting, and pagination to the query
func (q *BunQL) Apply(ctx context.Context, query *bun.SelectQuery) *bun.SelectQuery {
	// Apply filter
	if len(q.Filters.Filters) > 0 || len(q.Filters.Groups) > 0 {
		query = filter.ApplyFilterGroup(query, q.Filters)
	}

	// Apply sorting
	if len(q.Sort) > 0 {
		query = sorting.ApplySort(query, q.Sort)
	}

	// Apply pagination
	if q.Pagination != nil {
		query = pagination.ApplyPagination(query, q.Pagination)
	}

	return query
}

// ParseFromParams creates a BunQL instance from JSON/query parameters
func ParseFromParams(filterParam, sortParam string, page, limit int) (*BunQL, error) {
	ql := New()

	// Parse filter if provided
	if filterParam != "" {
		filters, err := filter.ParseFilters(filterParam)
		if err != nil {
			return nil, err
		}
		ql.WithFilters(filters)
	}

	// Parse sorting if provided
	if sortParam != "" {
		sort, err := sorting.ParseSort(sortParam)
		if err != nil {
			return nil, err
		}
		ql.WithSort(sort)
	}

	// Set up pagination if provided
	if page > 0 || limit > 0 {
		paging := &dto.Pagination{
			Page:  page,
			Limit: limit,
		}
		ql.WithPagination(paging)
	}

	return ql, nil
}

// GetPaginationMetadata calculates pagination metadata
func GetPaginationMetadata(p *dto.Pagination, totalCount int) map[string]interface{} {
	if p == nil || p.Limit <= 0 {
		return map[string]interface{}{
			"page":       1,
			"limit":      0,
			"totalPages": 1,
			"totalItems": totalCount,
		}
	}

	totalPages := totalCount / p.Limit
	if totalCount%p.Limit > 0 {
		totalPages++
	}

	currentPage := p.Page
	if currentPage < 1 {
		currentPage = 1
	}

	return map[string]interface{}{
		"page":       currentPage,
		"limit":      p.Limit,
		"totalPages": totalPages,
		"totalItems": totalCount,
	}
}
