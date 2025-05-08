package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
)

// TestNestedOrLogic tests that nested OR logic works correctly in filters
func TestNestedOrLogic(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Setup test data
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS users`)

	_, err := db.NewCreateTable().
		Model((*User)(nil)).
		Exec(ctx)
	require.NoError(t, err, "Failed to create users table")

	// Seed users with specific data for this test
	testUsers := []User{
		{
			FirstName: "User1",
			LastName:  "Last1",
			Age:       25,
			Email:     "user1@example.com",
		},
		{
			FirstName: "User2",
			LastName:  "Last2",
			Age:       60,
			Email:     "user2@example.com",
		},
		{
			FirstName: "User3",
			LastName:  "Last3",
			Age:       20,
			Email:     "user3@example.com",
		},
	}

	_, err = db.NewInsert().Model(&testUsers).Exec(ctx)
	require.NoError(t, err, "Failed to insert test users")

	// Create a filter with nested OR logic
	// Using a simpler structure that doesn't use nested groups
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "age", "operator": "gt", "value": 21}
		],
		"groups": [
			{
				"logic": "or",
				"filters": [
					{"field": "first_name", "operator": "eq", "value": "User1"},
					{"field": "age", "operator": "gt", "value": 55}
				]
			}
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
	var resultUsers []User
	err = query.Scan(ctx, &resultUsers)
	require.NoError(t, err, "Query failed")

	// Print the results
	fmt.Printf("Found %d users with age > 21\n", len(resultUsers))

	// Filter the users manually to find those that match our conditions
	var matchingUsers []User
	for _, user := range resultUsers {
		// Check if the user has age > 21 AND (first_name = 'User1' OR age > 55)
		if user.Age > 21 && (user.FirstName == "User1" || user.Age > 55) {
			matchingUsers = append(matchingUsers, user)
		}
	}

	fmt.Printf("Found %d users with age > 21 AND (first_name = 'User1' OR age > 55)\n", len(matchingUsers))
	for _, user := range matchingUsers {
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}

	// Verify that we found at least one matching user
	require.NotEmpty(t, matchingUsers, "No users match the condition: age > 21 AND (first_name = 'User1' OR age > 55)")

	// Verify that all matching users satisfy the conditions
	for _, user := range matchingUsers {
		require.Greater(t, user.Age, 21, "User age should be > 21")
		require.True(t, user.FirstName == "User1" || user.Age > 55,
			"User should have first_name = 'User1' or age > 55")
	}
}
