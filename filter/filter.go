package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fxnoob/bunql/dto"
	"github.com/fxnoob/bunql/operator"
	"github.com/uptrace/bun"
	"regexp"
	"strings"
)

func ParseFilters(jsonStr string) (dto.FilterGroup, error) {
	var group dto.FilterGroup
	err := json.Unmarshal([]byte(jsonStr), &group)
	if err != nil {
		return dto.FilterGroup{}, err
	}

	// Default to AND logic if not specified
	if group.Logic == "" {
		group.Logic = "and"
	}

	return group, nil
}

// ApplyFilterGroup applies a filter group to the query
func ApplyFilterGroup(query *bun.SelectQuery, group dto.FilterGroup) *bun.SelectQuery {
	if len(group.Filters) == 0 && len(group.Groups) == 0 {
		return query
	}

	// Get the logic from the group
	logic := strings.ToLower(group.Logic)
	if logic == "" {
		logic = "and"
	}

	// Apply the filter group
	return query.WhereGroup(logic, func(q *bun.SelectQuery) *bun.SelectQuery {
		// Apply all direct filters in this group
		for _, filter := range group.Filters {
			q = ApplyFilter(q, filter)
		}

		// Apply all nested filter groups
		for _, nestedGroup := range group.Groups {
			nestedLogic := strings.ToLower(nestedGroup.Logic)
			if nestedLogic == "" {
				nestedLogic = "and"
			}

			// Apply the nested group as a sub-group with the correct logic
			q = q.WhereGroup(nestedLogic, func(subq *bun.SelectQuery) *bun.SelectQuery {
				for _, filter := range nestedGroup.Filters {
					subq = ApplyFilter(subq, filter)
				}
				return subq
			})
		}
		return q
	})
}

// ApplyFilter applies a single filter to the query
func ApplyFilter(query *bun.SelectQuery, filter dto.Filter) *bun.SelectQuery {
	field := filter.Field
	op := operator.GetOperator(filter.Operator)
	value := filter.Value

	// Handle different operator
	switch op {
	case "=", "!=", ">", ">=", "<", "<=":
		// Handle date string values by using CAST for date columns
		if strValue, ok := value.(string); ok {
			// Check if this might be a date string (simple heuristic)
			if isDateString(strValue) {
				// Use CONVERT function for MSSQL to handle date comparison
				return query.Where(fmt.Sprintf("CONVERT(DATE, ?) %s CONVERT(DATE, ?)", op), bun.Ident(field), strValue)
			}
		}
		return query.Where(fmt.Sprintf("? %s ?", op), bun.Ident(field), value)
	case "LIKE":
		// Check if the value is a string
		if strValue, ok := value.(string); ok {
			// If the value doesn't already contain wildcards, add them
			if !strings.Contains(strValue, "%") {
				strValue = fmt.Sprintf("%%%s%%", strValue)
			}
			return query.Where("? LIKE ?", bun.Ident(field), strValue)
		}
		// If the value is not a string, use the default behavior
		likeValue := fmt.Sprintf("%%%v%%", value)
		return query.Where("? LIKE ?", bun.Ident(field), likeValue)
	case "IN":
		// Handle array values for IN operator
		return query.Where("? IN (?)", bun.Ident(field), bun.In(value))
	case "NOT IN":
		// Handle array values for NOT IN operator
		return query.Where("? NOT IN (?)", bun.Ident(field), bun.In(value))
	case "IS NULL":
		return query.Where("? IS NULL", bun.Ident(field))
	case "IS NOT NULL":
		return query.Where("? IS NOT NULL", bun.Ident(field))
	case "BETWEEN":
		// Handle array values for BETWEEN operator
		// The value should be an array or slice with two elements: [lowerBound, upperBound]
		if arr, ok := value.([]interface{}); ok && len(arr) == 2 {
			// Check if both values might be date strings
			if strVal1, ok1 := arr[0].(string); ok1 {
				if strVal2, ok2 := arr[1].(string); ok2 {
					if isDateString(strVal1) && isDateString(strVal2) {
						return query.Where("CONVERT(DATE, ?) BETWEEN CONVERT(DATE, ?) AND CONVERT(DATE, ?)",
							bun.Ident(field), strVal1, strVal2)
					}
				}
			}
			return query.Where("? BETWEEN ? AND ?", bun.Ident(field), arr[0], arr[1])
		}
		// If the value is not a valid array, return an error or default behavior
		return query.Where("? = ?", bun.Ident(field), value)
	default:
		// If operator not recognized, default to equality
		return query.Where("? = ?", bun.Ident(field), value)
	}
}

// validateFilterGroup validates a filter group and its nested filter
func validateFilterGroup(group dto.FilterGroup) error {
	logic := strings.ToLower(group.Logic)
	if logic != "and" && logic != "or" {
		return errors.New("filter group logic must be 'and' or 'or'")
	}

	// Validate individual filter
	for _, filter := range group.Filters {
		if err := validateFilter(filter); err != nil {
			return err
		}
	}

	// Validate nested groups
	for _, nestedGroup := range group.Groups {
		if err := validateFilterGroup(nestedGroup); err != nil {
			return err
		}
	}

	return nil
}

// validateFilter validates a single filter
func validateFilter(filter dto.Filter) error {
	if filter.Field == "" {
		return errors.New("filter field cannot be empty")
	}

	if !operator.IsValidOperator(filter.Operator) {
		return fmt.Errorf("invalid operator: %s", filter.Operator)
	}

	return nil
}

// isDateString checks if a string might be a date string
// This is a simple heuristic and might need to be adjusted based on your date formats
func isDateString(s string) bool {
	// Common date formats: YYYY-MM-DD, MM/DD/YYYY, DD-MM-YYYY, etc.
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),                  // YYYY-MM-DD
		regexp.MustCompile(`^\d{1,2}/\d{1,2}/\d{4}$`),              // MM/DD/YYYY
		regexp.MustCompile(`^\d{1,2}-\d{1,2}-\d{4}$`),              // DD-MM-YYYY
		regexp.MustCompile(`^\d{4}/\d{1,2}/\d{1,2}$`),              // YYYY/MM/DD
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`), // ISO format
	}

	for _, pattern := range datePatterns {
		if pattern.MatchString(s) {
			return true
		}
	}
	return false
}
