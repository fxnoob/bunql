package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
)

// TestOrLogic tests that OR logic works correctly in filters
func TestOrLogic(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Create a filter with OR logic
	filterJSON := `{
		"logic": "or",
		"filters": [
			{"field": "first_name", "operator": "like", "value": "User1"},
			{"field": "age", "operator": "gt", "value": 55}
		]
	}`

	// Parse the filter JSON
	ql, err := bunql.ParseFromParams(filterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse parameters")

	// Print the parsed filter for debugging
	fmt.Printf("Parsed Filters: %+v\n", ql.Filters)

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Print the results
	fmt.Printf("Found %d users with first_name like 'User1' OR age > 55\n", len(users))
	for _, user := range users {
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
		// Verify that each user matches at least one of the conditions
		require.True(t, user.FirstName == "User1" || user.Age > 55,
			"User should have first_name = 'User1' or age > 55")
	}
}
