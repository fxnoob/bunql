package e2e

import (
	"context"
	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSortParams(t *testing.T) {
	// Test with valid parameters
	sortJSON := bunql.ParseSortParams("last_name", "asc")
	require.Equal(t, `[{"field": "last_name", "dir": "asc"}]`, sortJSON)

	// Test with invalid direction (should default to "asc")
	sortJSON = bunql.ParseSortParams("age", "invalid")
	require.Equal(t, `[{"field": "age", "dir": "asc"}]`, sortJSON)

	// Test with empty sortby (should return empty string)
	sortJSON = bunql.ParseSortParams("", "asc")
	require.Equal(t, "", sortJSON)

	// Test with desc direction
	sortJSON = bunql.ParseSortParams("first_name", "desc")
	require.Equal(t, `[{"field": "first_name", "dir": "desc"}]`, sortJSON)
}

func TestParseSortParamsWithQuery(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Create a simple filter for users with age > 20
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "age", "operator": "gt", "value": 20}
		]
	}`

	// Use ParseSortParams to create the sort JSON
	sortJSON := bunql.ParseSortParams("last_name", "asc")

	// Parse the filter and sort JSON
	ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 5)
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
	for _, user := range users {
		require.Greater(t, user.Age, 20, "User age should be greater than 20")
	}
}