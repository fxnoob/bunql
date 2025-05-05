package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/fxnoob/bunql/dto"
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

// TestQueryWithCount demonstrates how to get the total count of records along with paginated results
func TestQueryWithCount(t *testing.T) {
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

	// Parse the filter and sort JSON with pagination (page 1, pageSize 5)
	ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 5)
	require.NoError(t, err, "Failed to parse parameters")

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))

	// Apply the BunQL filters, sort, and pagination, and get a count query
	mainQuery, countQuery := ql.ApplyWithCount(ctx, query)

	// Execute both queries and get the results along with the total count
	users, totalCount, err := bunql.ExecuteWithCount[User](ctx, mainQuery, countQuery)
	require.NoError(t, err, "Query execution failed")

	// Verify results
	fmt.Printf("Found %d users with age > 20 (total: %d)\n", len(users), totalCount)
	for _, user := range users {
		require.Greater(t, user.Age, 20, "User age should be greater than 20")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}

	// Calculate pagination metadata with base URI for prev/next links
	metadata := bunql.GetPaginationMetadata(ql.Pagination, totalCount, "https://api.example.com/users")
	fmt.Printf("Pagination metadata: %+v\n", metadata)
}

// TestBetweenOperator demonstrates the use of the between operator
func TestBetweenOperator(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Create a filter using the between operator for users with age between 25 and 40
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "age", "operator": "between", "value": [25, 40]}
		]
	}`

	// Create a simple sort by age
	sortJSON := `[{"field": "age", "dir": "asc"}]`

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
	fmt.Printf("Found %d users with age between 25 and 40\n", len(users))
	for _, user := range users {
		require.GreaterOrEqual(t, user.Age, 25, "User age should be greater than or equal to 25")
		require.LessOrEqual(t, user.Age, 40, "User age should be less than or equal to 40")
		fmt.Printf("User: %s %s, Age: %d\n", user.FirstName, user.LastName, user.Age)
	}
}

// TestPaginationMetadata tests the GetPaginationMetadata function
func TestPaginationMetadata(t *testing.T) {
	// Test case 1: Base URI with query parameters
	pagination := &dto.Pagination{
		Page:     2,
		PageSize: 10,
	}
	totalCount := 100
	baseURI := "https://api.example.com/users?filter=active&sort=name"

	metadata := bunql.GetPaginationMetadata(pagination, totalCount, baseURI)

	// Check that the metadata contains the expected values
	require.Equal(t, 2, metadata["page"])
	require.Equal(t, 10, metadata["pageSize"])
	require.Equal(t, 10, metadata["total"])
	require.Equal(t, 100, metadata["totalItems"])

	// Check that prev and next URLs contain the original query parameters
	prevURL, ok := metadata["prev"].(string)
	require.True(t, ok, "prev should be a string")
	require.Contains(t, prevURL, "filter=active")
	require.Contains(t, prevURL, "sort=name")
	require.Contains(t, prevURL, "page=1")
	require.Contains(t, prevURL, "pageSize=10")

	nextURL, ok := metadata["next"].(string)
	require.True(t, ok, "next should be a string")
	require.Contains(t, nextURL, "filter=active")
	require.Contains(t, nextURL, "sort=name")
	require.Contains(t, nextURL, "page=3")
	require.Contains(t, nextURL, "pageSize=10")

	// Test case 2: First page (prev should be omitted)
	pagination.Page = 1
	metadata = bunql.GetPaginationMetadata(pagination, totalCount, baseURI)

	// Check that prev is not included in the metadata
	_, hasPrev := metadata["prev"]
	require.False(t, hasPrev, "prev should not be included for the first page")

	// Check that next is included
	_, hasNext := metadata["next"]
	require.True(t, hasNext, "next should be included")

	// Test case 3: Last page (next should be omitted)
	pagination.Page = 10
	metadata = bunql.GetPaginationMetadata(pagination, totalCount, baseURI)

	// Check that prev is included
	_, hasPrev = metadata["prev"]
	require.True(t, hasPrev, "prev should be included for the last page")

	// Check that next is not included
	_, hasNext = metadata["next"]
	require.False(t, hasNext, "next should not be included for the last page")

	// Test case 4: No pagination (both prev and next should be omitted)
	metadata = bunql.GetPaginationMetadata(nil, totalCount, baseURI)

	// Check that neither prev nor next is included
	_, hasPrev = metadata["prev"]
	require.False(t, hasPrev, "prev should not be included when pagination is nil")

	_, hasNext = metadata["next"]
	require.False(t, hasNext, "next should not be included when pagination is nil")
}
