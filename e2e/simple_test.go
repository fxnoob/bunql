package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
)

// TestSimpleQuery demonstrates a simple query using bunql
func TestSimpleQuery(t *testing.T) {
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

	// Create a simple sort by last_name
	sortJSON := `[{"field": "last_name", "dir": "asc"}]`

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
	fmt.Printf("Found %d users with age > 20\n", len(users))
	for _, user := range users {
		require.Greater(t, user.Age, 20, "User age should be greater than 20")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}
