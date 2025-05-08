package e2e

import (
	"context"
	"fmt"
	"github.com/fxnoob/bunql"
	"github.com/fxnoob/bunql/filter"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestParseFilterParam demonstrates how to use the ParseFilterParam utility
func TestParseFilterParam(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Use the new ParseFilterParam utility to create a filter for users with age > 30
	filterGroup, err := filter.ParseFilterParam("age", "gt", 30, "and")
	require.NoError(t, err, "Failed to parse filter parameter")

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
	fmt.Printf("Found %d users with age > 30\n", len(users))
	for _, user := range users {
		require.Greater(t, user.Age, 30, "User age should be greater than 30")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}

// TestMultipleParseFilterParams demonstrates how to combine multiple filters created with ParseFilterParam
func TestMultipleParseFilterParams(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Create a filter for users with age > 25
	ageFilter, err := filter.ParseFilterParam("age", "gt", 25, "and")
	require.NoError(t, err, "Failed to parse age filter parameter")

	// Create a filter for users with first_name like "J"
	nameFilter, err := filter.ParseFilterParam("first_name", "like", "J", "and")
	require.NoError(t, err, "Failed to parse name filter parameter")

	// Combine the filters into a single filter group
	combinedFilter := ageFilter
	combinedFilter.Filters = append(combinedFilter.Filters, nameFilter.Filters...)

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

// TestParseFilterParamWithOrLogic demonstrates how to use the ParseFilterParam utility with OR logic
func TestParseFilterParamWithOrLogic(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Create a filter for users with age > 40 OR first_name like "A"
	ageFilter, err := filter.ParseFilterParam("age", "gt", 40, "or")
	require.NoError(t, err, "Failed to parse age filter parameter")

	// Create a filter for users with first_name like "A"
	nameFilter, err := filter.ParseFilterParam("first_name", "like", "A", "or")
	require.NoError(t, err, "Failed to parse name filter parameter")

	// Combine the filters into a single filter group
	combinedFilter := ageFilter
	combinedFilter.Filters = append(combinedFilter.Filters, nameFilter.Filters...)

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
