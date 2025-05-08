package dto

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GetPaginationMetadataOutput represents the output of pagination metadata
type GetPaginationMetadataOutput struct {
	Total     int     `json:"total"`
	Prev      *string `json:"prev"`
	Next      *string `json:"next"`
	TotalItem int     `json:"totalItem"`
}

// SortField represents a field to sorting by and the direction
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"dir"` // "asc" or "desc"
}

// FilterGroup represents a group of filter with a logical operator
type FilterGroup struct {
	Logic   string        `json:"logic"`   // "and" or "or"
	Filters []Filter      `json:"filters"` // List of filter
	Groups  []FilterGroup `json:"groups"`  // Nested filter groups
}

// Filter represents a single filter condition
type Filter struct {
	Field    string      `json:"field"`    // Field name to filter on
	Operator string      `json:"operator"` // Operator to use (eq, neq, gt, etc.)
	Value    interface{} `json:"value"`    // Value to compare against
}
