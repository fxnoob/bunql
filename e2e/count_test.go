package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type UserSession struct {
	bun.BaseModel `bun:"table:user_sessions,alias:user_sessions"`

	ID     int64 `bun:"id,pk,autoincrement"`
	UserID int64 `bun:"UserID"`
}

func TestCountQuery(t *testing.T) {
	// Get database connection
	db = GetDB()

	ctx := context.Background()

	// Drop the table if it exists
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS user_sessions`)
	require.NoError(t, err, "Failed to drop table")

	// Create the table
	_, err = db.NewCreateTable().
		Model((*UserSession)(nil)).
		Exec(ctx)
	require.NoError(t, err, "Failed to create table")

	// Insert some test data
	sessions := []UserSession{
		{UserID: 1},
		{UserID: 1},
		{UserID: 2},
	}
	_, err = db.NewInsert().Model(&sessions).Exec(ctx)
	require.NoError(t, err, "Failed to insert data")

	// Create a filter for UserID = 1
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "UserID", "operator": "eq", "value": 1}
		]
	}`

	// Parse the filter
	ql, err := bunql.ParseFromParams(filterJSON, "", 1, 10)
	require.NoError(t, err, "Failed to parse parameters")

	// Create a base query
	query := db.NewSelect().Model((*UserSession)(nil))

	// Apply the BunQL filters and get a count query
	mainQuery, countQuery := ql.ApplyWithCount(ctx, query)

	// Execute both queries and get the results along with the total count
	sessions, totalCount, err := bunql.ExecuteWithCount[UserSession](ctx, mainQuery, countQuery)
	require.NoError(t, err, "Query execution failed")

	// Verify results
	require.Equal(t, 2, totalCount, "Expected 2 sessions for UserID = 1")
	require.Len(t, sessions, 2, "Expected 2 sessions in the result")

	fmt.Printf("Found %d sessions for UserID = 1 (total: %d)\n", len(sessions), totalCount)
	for _, session := range sessions {
		fmt.Printf("Session ID: %d, UserID: %d\n", session.ID, session.UserID)
		require.Equal(t, int64(1), session.UserID, "All sessions should have UserID = 1")
	}
}
