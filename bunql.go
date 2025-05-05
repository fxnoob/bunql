package bunql

import (
	"context"
	"fmt"
	"github.com/fxnoob/bunql/dto"
	"github.com/fxnoob/bunql/filter"
	"github.com/fxnoob/bunql/pagination"
	"github.com/fxnoob/bunql/sorting"
	"github.com/uptrace/bun"
	"net/url"
	"strings"
)

type BunQL struct {
	Filters             dto.FilterGroup
	Sort                []dto.SortField
	Pagination          *dto.Pagination
	AllowedFilterFields []string
	AllowedSortFields   []string
}

// New creates a new BunQL instance
func New() *BunQL {
	return &BunQL{
		Filters: dto.FilterGroup{
			Logic:   "and",
			Filters: []dto.Filter{},
			Groups:  []dto.FilterGroup{},
		},
		Sort:                []dto.SortField{},
		Pagination:          nil,
		AllowedFilterFields: []string{},
		AllowedSortFields:   []string{},
	}
}

// NewWithAllowedFields creates a new BunQL instance with allowed fields for filtering and sorting
func NewWithAllowedFields(allowedFilterFields, allowedSortFields []string) *BunQL {
	return &BunQL{
		Filters: dto.FilterGroup{
			Logic:   "and",
			Filters: []dto.Filter{},
			Groups:  []dto.FilterGroup{},
		},
		Sort:                []dto.SortField{},
		Pagination:          nil,
		AllowedFilterFields: allowedFilterFields,
		AllowedSortFields:   allowedSortFields,
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

	// Print the query to console
	fmt.Println("Query:", query)

	return query
}

// ApplyWithCount applies all filter, sorting, and pagination to the query and returns both the query and a count query
func (q *BunQL) ApplyWithCount(ctx context.Context, query *bun.SelectQuery) (*bun.SelectQuery, *bun.SelectQuery) {
	// Create a copy of the query for counting
	countQuery := query.Clone()

	// Apply filter to both queries
	if len(q.Filters.Filters) > 0 || len(q.Filters.Groups) > 0 {
		query = filter.ApplyFilterGroup(query, q.Filters)
		countQuery = filter.ApplyFilterGroup(countQuery, q.Filters)
	}

	// Apply sorting only to the main query
	if len(q.Sort) > 0 {
		query = sorting.ApplySort(query, q.Sort)
	}

	// Apply pagination only to the main query
	if q.Pagination != nil {
		query = pagination.ApplyPagination(query, q.Pagination)
	}

	// Print the queries to console
	fmt.Println("Query:", query)
	fmt.Println("Count Query:", countQuery)

	return query, countQuery
}

// ParseFromParams creates a BunQL instance from JSON/query parameters
func ParseFromParams(filterParam, sortParam string, page, pageSize int) (*BunQL, error) {
	return ParseFromParamsWithAllowedFields(filterParam, sortParam, page, pageSize, nil, nil)
}

// ParseFromParamsWithAllowedFields creates a BunQL instance from JSON/query parameters with allowed fields for filtering and sorting
func ParseFromParamsWithAllowedFields(filterParam, sortParam string, page, pageSize int, allowedFilterFields, allowedSortFields []string) (*BunQL, error) {
	ql := NewWithAllowedFields(allowedFilterFields, allowedSortFields)

	// Parse filter if provided
	if filterParam != "" {
		filters, err := filter.ParseFilters(filterParam)
		if err != nil {
			return nil, err
		}

		// Validate filter fields if allowed fields are specified
		if len(ql.AllowedFilterFields) > 0 {
			if err := validateFilterFields(filters, ql.AllowedFilterFields); err != nil {
				return nil, err
			}
		}

		ql.WithFilters(filters)
	}

	// Parse sorting if provided
	if sortParam != "" {
		sort, err := sorting.ParseSort(sortParam)
		if err != nil {
			return nil, err
		}

		// Validate sort fields if allowed fields are specified
		if len(ql.AllowedSortFields) > 0 {
			if err := validateSortFields(sort, ql.AllowedSortFields); err != nil {
				return nil, err
			}
		}

		ql.WithSort(sort)
	}

	// Set up pagination if provided
	if page > 0 || pageSize > 0 {
		paging := &dto.Pagination{
			Page:     page,
			PageSize: pageSize,
		}
		ql.WithPagination(paging)
	}

	return ql, nil
}

// validateFilterFields validates that all filter fields are in the list of allowed fields
func validateFilterFields(group dto.FilterGroup, allowedFields []string) error {
	// Validate all direct filters in this group
	for _, filter := range group.Filters {
		if !contains(allowedFields, filter.Field) {
			return fmt.Errorf("filter field '%s' is not allowed", filter.Field)
		}
	}

	// Validate all nested filter groups
	for _, nestedGroup := range group.Groups {
		if err := validateFilterFields(nestedGroup, allowedFields); err != nil {
			return err
		}
	}

	return nil
}

// validateSortFields validates that all sort fields are in the list of allowed fields
func validateSortFields(sortFields []dto.SortField, allowedFields []string) error {
	for _, sort := range sortFields {
		if !contains(allowedFields, sort.Field) {
			return fmt.Errorf("sort field '%s' is not allowed", sort.Field)
		}
	}

	return nil
}

// contains checks if a string is in a slice of strings
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetPaginationMetadata calculates pagination metadata and generates prev/next URLs
func GetPaginationMetadata(p *dto.Pagination, totalCount int, baseURI string) map[string]interface{} {
	if p == nil || p.PageSize <= 0 {
		return map[string]interface{}{
			"page":       1,
			"pageSize":   0,
			"total":      1,
			"totalItems": totalCount,
		}
	}

	total := totalCount / p.PageSize
	if totalCount%p.PageSize > 0 {
		total++
	}

	currentPage := p.Page
	if currentPage < 1 {
		currentPage = 1
	}

	// Parse the baseURI to extract any existing query parameters
	var baseURL string
	var queryParams map[string][]string

	// Check if the baseURI contains a query string
	parts := strings.Split(baseURI, "?")
	baseURL = parts[0]

	// If there's a query string, parse it
	if len(parts) > 1 && parts[1] != "" {
		parsedURL, err := url.Parse(baseURI)
		if err == nil {
			queryParams = parsedURL.Query()
		} else {
			queryParams = make(map[string][]string)
		}
	} else {
		queryParams = make(map[string][]string)
	}

	// Generate prev and next URLs
	var prevURL, nextURL string

	if currentPage > 1 {
		// Create a copy of the query parameters for the prev URL
		prevParams := make(map[string][]string)
		for k, v := range queryParams {
			prevParams[k] = v
		}
		prevParams["page"] = []string{fmt.Sprintf("%d", currentPage-1)}
		prevParams["pageSize"] = []string{fmt.Sprintf("%d", p.PageSize)}

		// Build the query string
		var queryStr string
		first := true
		for k, values := range prevParams {
			for _, v := range values {
				if first {
					queryStr += "?"
					first = false
				} else {
					queryStr += "&"
				}
				queryStr += url.QueryEscape(k) + "=" + url.QueryEscape(v)
			}
		}

		prevURL = baseURL + queryStr
	}

	if currentPage < total {
		// Create a copy of the query parameters for the next URL
		nextParams := make(map[string][]string)
		for k, v := range queryParams {
			nextParams[k] = v
		}
		nextParams["page"] = []string{fmt.Sprintf("%d", currentPage+1)}
		nextParams["pageSize"] = []string{fmt.Sprintf("%d", p.PageSize)}

		// Build the query string
		var queryStr string
		first := true
		for k, values := range nextParams {
			for _, v := range values {
				if first {
					queryStr += "?"
					first = false
				} else {
					queryStr += "&"
				}
				queryStr += url.QueryEscape(k) + "=" + url.QueryEscape(v)
			}
		}

		nextURL = baseURL + queryStr
	}

	// Create the result map
	result := map[string]interface{}{
		"page":       currentPage,
		"pageSize":   p.PageSize,
		"total":      total,
		"totalItems": totalCount,
	}

	// Only include prev and next if they have values
	if prevURL != "" {
		result["prev"] = prevURL
	}
	if nextURL != "" {
		result["next"] = nextURL
	}

	return result
}

// ExecuteWithCount executes both the main query and the count query, and returns the results along with the total count
func ExecuteWithCount[T any](ctx context.Context, query, countQuery *bun.SelectQuery) ([]T, int, error) {
	// Execute the count query
	count, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	// Execute the main query
	var results []T
	if err := query.Scan(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to execute main query: %w", err)
	}

	return results, count, nil
}
