package e2e

import (
	"context"
	"fmt"
	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestParseFilterParams demonstrates how to use the ParseFilterParams utility
func TestParseFilterParams(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Use the new ParseFilterParams utility to create a filter JSON string for users with age > 30
	filterJSON, err := bunql.ParseFilterParams("age", "gt", 30, "and")
	require.NoError(t, err, "Failed to parse filter parameters")

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
	fmt.Printf("Found %d users with age > 30\n", len(users))
	for _, user := range users {
		require.Greater(t, user.Age, 30, "User age should be greater than 30")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}

// TestCombineParseFilterParams demonstrates how to combine multiple filters created with ParseFilterParams
func TestCombineParseFilterParams(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Create a filter JSON string for users with age > 25
	ageFilterJSON, err := bunql.ParseFilterParams("age", "gt", 25, "and")
	require.NoError(t, err, "Failed to parse age filter parameters")

	// Create a filter JSON string for users with first_name like "J"
	nameFilterJSON, err := bunql.ParseFilterParams("first_name", "like", "J", "and")
	require.NoError(t, err, "Failed to parse name filter parameters")

	// Parse the individual filter JSON strings to get filter groups
	ageFilterGroup, err := bunql.ParseFromParams(ageFilterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse age filter parameters")

	nameFilterGroup, err := bunql.ParseFromParams(nameFilterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse name filter parameters")

	// Combine the filters into a single filter group
	combinedFilter := ageFilterGroup.Filters
	combinedFilter.Filters = append(combinedFilter.Filters, nameFilterGroup.Filters.Filters...)

	// Create a BunQL instance with the combined filter
	ql := bunql.New().WithFilters(combinedFilter)

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	require.NoError(t, err, "Query failed")

	// Verify results
	fmt.Printf("Found %d users with age > 25 and first_name like 'J'\n", len(users))
	for _, user := range users {
		require.Greater(t, user.Age, 25, "User age should be greater than 25")
		require.Contains(t, user.FirstName, "J", "User first name should contain 'J'")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}

// TestOrLogicWithParseFilterParams demonstrates how to use the ParseFilterParams utility with OR logic
func TestOrLogicWithParseFilterParams(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Create a filter JSON string for users with age > 40 OR first_name like "A"
	ageFilterJSON, err := bunql.ParseFilterParams("age", "gt", 40, "or")
	require.NoError(t, err, "Failed to parse age filter parameters")

	// Create a filter JSON string for users with first_name like "A"
	nameFilterJSON, err := bunql.ParseFilterParams("first_name", "like", "A", "or")
	require.NoError(t, err, "Failed to parse name filter parameters")

	// Parse the individual filter JSON strings to get filter groups
	ageFilterGroup, err := bunql.ParseFromParams(ageFilterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse age filter parameters")

	nameFilterGroup, err := bunql.ParseFromParams(nameFilterJSON, "", 0, 0)
	require.NoError(t, err, "Failed to parse name filter parameters")

	// Combine the filters into a single filter group
	combinedFilter := ageFilterGroup.Filters
	combinedFilter.Filters = append(combinedFilter.Filters, nameFilterGroup.Filters.Filters...)

	// Create a BunQL instance with the combined filter
	ql := bunql.New().WithFilters(combinedFilter)

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
