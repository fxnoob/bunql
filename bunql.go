package bunql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fxnoob/bunql/dto"
	"github.com/fxnoob/bunql/filter"
	"github.com/fxnoob/bunql/pagination"
	"github.com/fxnoob/bunql/sorting"
	"github.com/uptrace/bun"
	"net/url"
	"strings"
)

// Filter is a re-export of dto.Filter to make it accessible directly from the bunql package
type Filter = dto.Filter

// PaginationMetadataOutput is an alias for dto.GetPaginationMetadataOutput
type PaginationMetadataOutput = dto.GetPaginationMetadataOutput

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
	// Apply the filters, sorting, and pagination to the main query
	mainQuery := q.Apply(ctx, query)

	// For the count query, only apply the filters
	countQuery := query
	if len(q.Filters.Filters) > 0 || len(q.Filters.Groups) > 0 {
		countQuery = filter.ApplyFilterGroup(countQuery, q.Filters)
	}

	// Print the queries to console
	fmt.Println("Main Query:", mainQuery)
	fmt.Println("Count Query:", countQuery)

	return mainQuery, countQuery
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
func GetPaginationMetadata(p *dto.Pagination, totalCount int, baseURI string) PaginationMetadataOutput {
	if p == nil || p.PageSize <= 0 {
		return PaginationMetadataOutput{
			Total:     1,
			TotalItem: totalCount,
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
	var prevURL, nextURL *string

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

		prevURLStr := baseURL + queryStr
		prevURL = &prevURLStr
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

		nextURLStr := baseURL + queryStr
		nextURL = &nextURLStr
	}

	// Create the result using the type alias
	result := PaginationMetadataOutput{
		Total:     total,
		Prev:      prevURL,
		Next:      nextURL,
		TotalItem: totalCount,
	}

	return result
}

// ParseSortParams creates a sort JSON string from sortby and sortDirection parameters
// sortby is the field name to sort by
// sortDirection is the sort direction, which can be "asc" or "desc" (defaults to "asc" if invalid)
func ParseSortParams(sortby, sortDirection string) string {
	// Default to "asc" if sortDirection is not valid
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "asc"
	}

	// Return empty string if sortby is empty
	if sortby == "" {
		return ""
	}

	// Create the sort JSON string
	return fmt.Sprintf(`[{"field": "%s", "dir": "%s"}]`, sortby, sortDirection)
}

// ParseFilterParams creates a filter JSON string from field, operator, and value parameters
// field is the field name to filter on
// operator is the operator to use (eq, neq, gt, etc.)
// value is the value to compare against
// logic is the logic to use for the filter group ("and" or "or", defaults to "and" if empty or invalid)
func ParseFilterParams(field, operator string, value interface{}, logic string) (string, error) {
	// Use the filter package's ParseFilterParam function
	filterGroup, err := filter.ParseFilterParam(field, operator, value, logic)
	if err != nil {
		return "", err
	}

	// Convert the filter group to JSON
	jsonBytes, err := json.Marshal(filterGroup)
	if err != nil {
		return "", fmt.Errorf("failed to marshal filter group to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// ParseMultipleFilterParams creates a filter JSON string from a slice of Filter DTOs
// filters is a slice of Filter DTOs, each containing a field, operator, and value
// logic is the logic to use for the filter group ("and" or "or", defaults to "and" if empty or invalid)
func ParseMultipleFilterParams(filters []Filter, logic string) (string, error) {
	// Default to "and" logic if not specified or invalid
	logic = strings.ToLower(logic)
	if logic != "and" && logic != "or" {
		logic = "and"
	}

	// Create a filter group
	filterGroup := dto.FilterGroup{
		Logic:   logic,
		Filters: []dto.Filter{},
		Groups:  []dto.FilterGroup{},
	}

	// Add each filter to the group
	for _, f := range filters {
		// Use the filter package's ParseFilterParam function to create a filter
		singleFilterGroup, err := filter.ParseFilterParam(f.Field, f.Operator, f.Value, logic)
		if err != nil {
			return "", err
		}

		// Add the filter to the group
		filterGroup.Filters = append(filterGroup.Filters, singleFilterGroup.Filters...)
	}

	// Convert the filter group to JSON
	jsonBytes, err := json.Marshal(filterGroup)
	if err != nil {
		return "", fmt.Errorf("failed to marshal filter group to JSON: %w", err)
	}

	return string(jsonBytes), nil
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
