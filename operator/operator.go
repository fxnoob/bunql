package operator

import "strings"

// Known operator map
var operatorMap = map[string]string{
	"eq":        "=",
	"neq":       "!=",
	"gt":        ">",
	"gte":       ">=",
	"lt":        "<",
	"lte":       "<=",
	"like":      "LIKE",
	"in":        "IN",
	"notin":     "NOT IN",
	"isnull":    "IS NULL",
	"isnotnull": "IS NOT NULL",
}

// GetOperator returns the SQL operator for a given operator name
func GetOperator(op string) string {
	op = strings.ToLower(op)
	if sqlOp, ok := operatorMap[op]; ok {
		return sqlOp
	}
	return "="
}

// IsValidOperator checks if an operator is valid
func IsValidOperator(op string) bool {
	op = strings.ToLower(op)
	_, ok := operatorMap[op]
	return ok
}

// GetSupportedOperators returns a list of all supported operator
func GetSupportedOperators() []string {
	operators := make([]string, 0, len(operatorMap))
	for op := range operatorMap {
		operators = append(operators, op)
	}
	return operators
}
