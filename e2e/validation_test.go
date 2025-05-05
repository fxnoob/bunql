package e2e

import (
	"context"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
)

// TestFieldValidation tests the validation of allowed fields for filtering and sorting
func TestFieldValidation(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Define allowed fields for filtering and sorting
	allowedFilterFields := []string{"age", "first_name"}
	allowedSortFields := []string{"last_name", "age"}

	// Test 1: Valid filter and sort fields
	t.Run("Valid fields", func(t *testing.T) {
		// Create a filter using only allowed fields
		filterJSON := `{
			"logic": "and",
			"filters": [
				{"field": "age", "operator": "gt", "value": 20}
			]
		}`

		// Create a sort using only allowed fields
		sortJSON := `[{"field": "last_name", "dir": "asc"}]`

		// Parse the filter and sort JSON with allowed fields
		ql, err := bunql.ParseFromParamsWithAllowedFields(filterJSON, sortJSON, 1, 5, allowedFilterFields, allowedSortFields)
		require.NoError(t, err, "Failed to parse parameters with valid fields")

		// Create a base query
		query := db.NewSelect().Model((*User)(nil))

		// Apply the BunQL filters, sort, and pagination
		query = ql.Apply(ctx, query)

		// Execute the query
		var users []User
		err = query.Scan(ctx, &users)
		require.NoError(t, err, "Query failed")
	})

	// Test 2: Invalid filter field
	t.Run("Invalid filter field", func(t *testing.T) {
		// Create a filter using a disallowed field
		filterJSON := `{
			"logic": "and",
			"filters": [
				{"field": "email", "operator": "like", "value": "example"}
			]
		}`

		// Create a sort using only allowed fields
		sortJSON := `[{"field": "last_name", "dir": "asc"}]`

		// Parse the filter and sort JSON with allowed fields
		_, err := bunql.ParseFromParamsWithAllowedFields(filterJSON, sortJSON, 1, 5, allowedFilterFields, allowedSortFields)
		require.Error(t, err, "Should fail with disallowed filter field")
		require.Contains(t, err.Error(), "filter field 'email' is not allowed")
	})

	// Test 3: Invalid sort field
	t.Run("Invalid sort field", func(t *testing.T) {
		// Create a filter using only allowed fields
		filterJSON := `{
			"logic": "and",
			"filters": [
				{"field": "age", "operator": "gt", "value": 20}
			]
		}`

		// Create a sort using a disallowed field
		sortJSON := `[{"field": "email", "dir": "asc"}]`

		// Parse the filter and sort JSON with allowed fields
		_, err := bunql.ParseFromParamsWithAllowedFields(filterJSON, sortJSON, 1, 5, allowedFilterFields, allowedSortFields)
		require.Error(t, err, "Should fail with disallowed sort field")
		require.Contains(t, err.Error(), "sort field 'email' is not allowed")
	})

	// Test 4: Nested filter group with invalid field
	t.Run("Nested filter group with invalid field", func(t *testing.T) {
		// Create a filter with a nested group using a disallowed field
		filterJSON := `{
			"logic": "and",
			"filters": [
				{"field": "age", "operator": "gt", "value": 20}
			],
			"groups": [
				{
					"logic": "or",
					"filters": [
						{"field": "email", "operator": "like", "value": "example"}
					]
				}
			]
		}`

		// Create a sort using only allowed fields
		sortJSON := `[{"field": "last_name", "dir": "asc"}]`

		// Parse the filter and sort JSON with allowed fields
		_, err := bunql.ParseFromParamsWithAllowedFields(filterJSON, sortJSON, 1, 5, allowedFilterFields, allowedSortFields)
		require.Error(t, err, "Should fail with disallowed filter field in nested group")
		require.Contains(t, err.Error(), "filter field 'email' is not allowed")
	})
}
