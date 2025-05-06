package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/fxnoob/bunql/dto"
	"github.com/stretchr/testify/require"
)

// TestNotInOperator demonstrates the use of the notin operator
func TestNotInOperator(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Create a filter using the notin operator for users with IDs not in [1, 2]
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "id", "operator": "notin", "value": [1, 2]}
		]
	}`

	// Create a simple sort by ID
	sortJSON := `[{"field": "id", "dir": "asc"}]`

	// Parse the filter and sort JSON
	ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 10)
	require.NoError(t, err, "Failed to parse parameters")

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters, sort, and pagination
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Verify results
	fmt.Printf("Found %d users with ID not in [1, 2]\n", len(users))
	for _, user := range users {
		require.NotContains(t, []int64{1, 2}, user.ID, "User ID should not be in [1, 2]")
		fmt.Printf("User: %s %s, ID: %d\n", user.FirstName, user.LastName, user.ID)
	}
}

// TestArrayNotInOperator demonstrates handling an array format similar to the issue description
func TestArrayNotInOperator(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Create a filter using the notin operator for ID not in [1, 2]
	// This is similar to the format in the issue description
	filterJSON := `[{"field": "id", "operator": "notin", "value": [1,2]}]`

	// Parse the filter JSON
	// Note: The filter JSON format is an array of filters, not a filter group
	// We need to convert it to a filter group
	var filters []map[string]interface{}
	err = json.Unmarshal([]byte(filterJSON), &filters)
	require.NoError(t, err, "Failed to parse filter JSON")

	// Convert to a filter group
	filterGroup := dto.FilterGroup{
		Logic:   "and",
		Filters: make([]dto.Filter, 0),
	}

	for _, f := range filters {
		filter := dto.Filter{
			Field:    f["field"].(string),
			Operator: f["operator"].(string),
			Value:    f["value"],
		}
		filterGroup.Filters = append(filterGroup.Filters, filter)
	}

	// Create a BunQL instance with the filter
	ql := bunql.New().WithFilters(filterGroup)

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Verify results
	fmt.Printf("Found %d users with ID not in [1, 2]\n", len(users))
	for _, user := range users {
		require.NotContains(t, []int64{1, 2}, user.ID, "User ID should not be in [1, 2]")
	}
}
