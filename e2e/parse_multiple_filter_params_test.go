package e2e

import (
	"context"
	"fmt"
	"github.com/fxnoob/bunql"
	"github.com/fxnoob/bunql/dto"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestParseMultipleFilterParams demonstrates how to use the ParseMultipleFilterParams utility
func TestParseMultipleFilterParams(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Use the new ParseMultipleFilterParams utility to create a filter JSON string for users with age > 30 and first_name like "J"
	filters := []dto.Filter{
		{Field: "age", Operator: "gt", Value: 30},
		{Field: "first_name", Operator: "like", Value: "J"},
	}
	filterJSON, err := bunql.ParseMultipleFilterParams(filters, "and")
	require.NoError(t, err, "Failed to parse multiple filter parameters")

	// Create a BunQL instance from the filter JSON string
	ql, err := bunql.ParseFromParams(filterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse parameters")

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Verify results
	fmt.Printf("Found %d users with age > 30 and first_name like 'J'\n", len(users))
	for _, user := range users {
		require.Greater(t, user.Age, 30, "User age should be greater than 30")
		require.Contains(t, user.FirstName, "J", "User first name should contain 'J'")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}

// TestParseMultipleFilterParamsWithOrLogic demonstrates how to use the ParseMultipleFilterParams utility with OR logic
func TestParseMultipleFilterParamsWithOrLogic(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Use the new ParseMultipleFilterParams utility to create a filter JSON string for users with age > 40 OR first_name like "A"
	filters := []dto.Filter{
		{Field: "age", Operator: "gt", Value: 40},
		{Field: "first_name", Operator: "like", Value: "A"},
	}
	filterJSON, err := bunql.ParseMultipleFilterParams(filters, "or")
	require.NoError(t, err, "Failed to parse multiple filter parameters")

	// Create a BunQL instance from the filter JSON string
	ql, err := bunql.ParseFromParams(filterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse parameters")

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Verify results
	fmt.Printf("Found %d users with age > 40 OR first_name like 'A'\n", len(users))
	for _, user := range users {
		// Either age > 40 OR first_name contains "A"
		isValid := user.Age > 40 || (user.FirstName != "" && (user.FirstName[0] == 'A' || user.FirstName[0] == 'a'))
		require.True(t, isValid, "User should have age > 40 OR first_name like 'A'")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}
