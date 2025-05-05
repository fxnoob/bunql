package sorting

import (
	"encoding/json"
	"fmt"
	"github.com/fxnoob/bunql/dto"
	"github.com/uptrace/bun"
	"strings"
)

func ParseSort(jsonStr string) ([]dto.SortField, error) {
	var sortFields []dto.SortField
	err := json.Unmarshal([]byte(jsonStr), &sortFields)
	if err != nil {
		return nil, err
	}

	// Validate and normalize directions
	for i := range sortFields {
		dir := strings.ToLower(sortFields[i].Direction)
		if dir != "asc" && dir != "desc" {
			sortFields[i].Direction = "asc" // Default to ascending
		} else {
			sortFields[i].Direction = dir
		}
	}

	return sortFields, nil
}

// ApplySort applies sorting to the query
func ApplySort(query *bun.SelectQuery, sortFields []dto.SortField) *bun.SelectQuery {
	for _, sort := range sortFields {
		orderExpr := fmt.Sprintf("%s %s", sort.Field, strings.ToUpper(sort.Direction))
		query = query.OrderExpr(orderExpr)
	}
	return query
}
