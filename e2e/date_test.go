package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// Event is a test model with a date column
type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID          int64     `bun:"id,pk,autoincrement"`
	Title       string    `bun:"title"`
	Description string    `bun:"description"`
	EventDate   time.Time `bun:"event_date"`
	IsActive    bool      `bun:"is_active"`
}

// TestDateFiltering tests filtering on date columns
func TestDateFiltering(t *testing.T) {
	// Get database connection
	db, err := GetDB()
	require.NoError(t, err, "Failed to connect to database")

	ctx := context.Background()

	// Drop if exists, then create
	_, _ = db.ExecContext(ctx, `IF OBJECT_ID('events', 'U') IS NOT NULL DROP TABLE events`)

	// Create the events table
	_, err = db.NewCreateTable().
		Model((*Event)(nil)).
		Exec(ctx)
	require.NoError(t, err, "Failed to create events table")

	// Create some test events with different dates
	events := []Event{
		{
			Title:       "Past Event 1",
			Description: "An event in the past",
			EventDate:   time.Now().AddDate(0, -1, 0), // 1 month ago
			IsActive:    true,
		},
		{
			Title:       "Past Event 2",
			Description: "Another event in the past",
			EventDate:   time.Now().AddDate(0, -2, 0), // 2 months ago
			IsActive:    true,
		},
		{
			Title:       "Future Event 1",
			Description: "An event in the future",
			EventDate:   time.Now().AddDate(0, 1, 0), // 1 month in the future
			IsActive:    true,
		},
		{
			Title:       "Future Event 2",
			Description: "Another event in the future",
			EventDate:   time.Now().AddDate(0, 2, 0), // 2 months in the future
			IsActive:    true,
		},
	}

	// Insert the events
	_, err = db.NewInsert().Model(&events).Exec(ctx)
	require.NoError(t, err, "Failed to insert events")

	// Test 1: Filter for events after today (using string date)
	t.Run("Filter for events after today", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		filterJSON := fmt.Sprintf(`{
			"logic": "and",
			"filters": [
				{"field": "event_date", "operator": "gt", "value": "%s"}
			]
		}`, today)

		// Parse the filter JSON
		ql, err := bunql.ParseFromParams(filterJSON, "", 1, 10)
		require.NoError(t, err, "Failed to parse parameters")

		// Create a base query
		query := db.NewSelect().Model((*Event)(nil))

		// Apply the BunQL filters
		query = ql.Apply(ctx, query)

		// Execute the query
		var results []Event
		err = query.Scan(ctx, &results)
		require.NoError(t, err, "Query failed")

		// Verify results - should only include future events
		require.Equal(t, 2, len(results), "Should find 2 future events")
		for _, event := range results {
			require.True(t, event.EventDate.After(time.Now().Truncate(24*time.Hour)),
				"Event date should be in the future")
		}
	})

	// Test 2: Filter for events between two dates (using string dates)
	t.Run("Filter for events between two dates", func(t *testing.T) {
		startDate := time.Now().AddDate(0, -1, -15).Format("2006-01-02") // 1.5 months ago
		endDate := time.Now().AddDate(0, 1, 15).Format("2006-01-02")     // 1.5 months in the future

		filterJSON := fmt.Sprintf(`{
			"logic": "and",
			"filters": [
				{"field": "event_date", "operator": "between", "value": ["%s", "%s"]}
			]
		}`, startDate, endDate)

		// Parse the filter JSON
		ql, err := bunql.ParseFromParams(filterJSON, "", 1, 10)
		require.NoError(t, err, "Failed to parse parameters")

		// Create a base query
		query := db.NewSelect().Model((*Event)(nil))

		// Apply the BunQL filters
		query = ql.Apply(ctx, query)

		// Execute the query
		var results []Event
		err = query.Scan(ctx, &results)
		require.NoError(t, err, "Query failed")

		// Verify results - should include events within the date range
		require.Equal(t, 2, len(results), "Should find 2 events within the date range")
	})

	// Test 3: Filter for events on a specific date (using string date)
	t.Run("Filter for events on a specific date", func(t *testing.T) {
		// Use the date of the first future event
		specificDate := events[2].EventDate.Format("2006-01-02")

		filterJSON := fmt.Sprintf(`{
			"logic": "and",
			"filters": [
				{"field": "event_date", "operator": "eq", "value": "%s"}
			]
		}`, specificDate)

		// Parse the filter JSON
		ql, err := bunql.ParseFromParams(filterJSON, "", 1, 10)
		require.NoError(t, err, "Failed to parse parameters")

		// Create a base query
		query := db.NewSelect().Model((*Event)(nil))

		// Apply the BunQL filters
		query = ql.Apply(ctx, query)

		// Execute the query
		var results []Event
		err = query.Scan(ctx, &results)
		require.NoError(t, err, "Query failed")

		// Verify results - should include only the event on the specific date
		require.Equal(t, 1, len(results), "Should find 1 event on the specific date")
		require.Equal(t, "Future Event 1", results[0].Title, "Should find the correct event")
	})

	// Test 4: Filter for events with a date in MM/DD/YYYY format
	t.Run("Filter with MM/DD/YYYY date format", func(t *testing.T) {
		// Use the date of the first future event in MM/DD/YYYY format
		futureEventDate := events[2].EventDate
		mmddyyyyDate := fmt.Sprintf("%02d/%02d/%04d",
			futureEventDate.Month(), futureEventDate.Day(), futureEventDate.Year())

		filterJSON := fmt.Sprintf(`{
			"logic": "and",
			"filters": [
				{"field": "event_date", "operator": "eq", "value": "%s"}
			]
		}`, mmddyyyyDate)

		// Parse the filter JSON
		ql, err := bunql.ParseFromParams(filterJSON, "", 1, 10)
		require.NoError(t, err, "Failed to parse parameters")

		// Create a base query
		query := db.NewSelect().Model((*Event)(nil))

		// Apply the BunQL filters
		query = ql.Apply(ctx, query)

		// Execute the query
		var results []Event
		err = query.Scan(ctx, &results)
		require.NoError(t, err, "Query failed")

		// Verify results - should include only the event on the specific date
		require.Equal(t, 1, len(results), "Should find 1 event on the specific date")
		require.Equal(t, "Future Event 1", results[0].Title, "Should find the correct event")
	})
}
