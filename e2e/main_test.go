package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fxnoob/bunql"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/extra/bundebug"
	"math/rand"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mssqldialect"
)

func GetDB() (*bun.DB, error) {
	dsn := "sqlserver://sa:YourStrong@Passw0rd@localhost:1433?database=master&encrypt=disable"
	sqldb, err := sql.Open("sqlserver", dsn)
	if err != nil {
		fmt.Println("DATABASE CONNECTION ERROR:", err)
		return nil, err
	}

	db := bun.NewDB(sqldb, mssqldialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(false),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return db, nil
}

func TestCreateAndSeedUsers(t *testing.T) {
	db, _ := GetDB()
	ctx := context.Background()
	// Drop if exists, then create
	_, _ = db.ExecContext(ctx, `IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users`)

	_, err := db.NewCreateTable().
		Model((*User)(nil)).
		Exec(ctx)
	if err != nil {
		panic(err)
	}

	// Seed 10 users
	users := make([]User, 10)
	for i := range users {
		users[i] = User{
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
			Age:       rand.Intn(50) + 18,
			Email:     fmt.Sprintf("user%d@example.com", i),
		}
	}

	_, err = db.NewInsert().Model(&users).Exec(ctx)
	require.NoError(t, err)

	// Fetch back and verify
	var out []User
	err = db.NewSelect().Model(&out).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, out, 10)
}

func TestFiltersAndSorting(t *testing.T) {
	ctx := context.Background()
	db, _ := GetDB()
	// Sample query filter JSON
	filterJSON := `{
		"logic": "and",
		"filters": [
			{"field": "age", "operator": "gt", "value": 21}
		],
		"groups": [
			{
				"logic": "or",
				"filters": [
					{"field": "first_name", "operator": "like", "value": "User1"},
					{"field": "age", "operator": "gt", "value": 55}
				]
			}
		]
	}`

	// Sample sort JSON
	sortJSON := `[
		{"field": "age", "dir": "desc"},
		{"field": "last_name", "dir": "asc"}
	]`

	// Parse the filter and sort JSON
	ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 10)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse params: %v", err))
	}

	// Print the filter JSON and the parsed filter object
	fmt.Println("Filter JSON:", filterJSON)
	fmt.Printf("Parsed Filters: %+v\n", ql.Filters)

	// Create a base query
	query := db.NewSelect().Model((*User)(nil))
	// Apply the BunQL filters, sort, and pagination
	query = ql.Apply(ctx, query)

	// Execute the query
	var users []User
	err = query.Scan(ctx, &users)
	if err != nil {
		panic(fmt.Sprintf("Query failed: %v", err))
	}

	// Print the results
	fmt.Println("Query results:")
	userJSON, _ := json.MarshalIndent(users, "", "  ")
	fmt.Println(string(userJSON))
	count, err := db.NewSelect().Model((*User)(nil)).Count(ctx)
	if err != nil {
		panic(fmt.Sprintf("Count failed: %v", err))
	}

	// Get pagination metadata with base URI for prev/next links
	meta := bunql.GetPaginationMetadata(ql.Pagination, count, "https://api.example.com/users")
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	fmt.Println("Pagination metadata:")
	fmt.Println(string(metaJSON))
}
