# BunQL

BunQL is a Go library that provides a structured way to build SQL queries with filtering, sorting, and pagination capabilities for the [Bun SQL toolkit](https://github.com/uptrace/bun).

## Features

- **Dynamic Filtering**: Create complex filter conditions from JSON configurations
- **Flexible Sorting**: Sort by multiple fields with customizable directions
- **Pagination**: Easily paginate results with metadata
- **Total Count**: Get the total count of records alongside paginated results
- **JSON-based Configuration**: Define filters and sorting using simple JSON structures
- **Fluent API**: Use method chaining for a clean and readable query building experience
- **Field Validation**: Validate that only allowed fields are used for filtering and sorting

## Installation

```bash
go get github.com/fxnoob/bunql
```

## Prerequisites

- Go 1.23.0 or later
- Docker and Docker Compose (for running the MS SQL Server database for tests)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/fxnoob/bunql"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
    "github.com/uptrace/bun/driver/sqliteshim"
    "database/sql"
)

type User struct {
    bun.BaseModel `bun:"table:users,alias:u"`

    ID        int64  `bun:"id,pk,autoincrement"`
    FirstName string `bun:"first_name"`
    LastName  string `bun:"last_name"`
    Email     string `bun:"email"`
    Age       int    `bun:"age"`
    Active    bool   `bun:"active"`
}

func main() {
    // Connect to database
    sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
    if err != nil {
        panic(err)
    }
    defer sqldb.Close()

    db := bun.NewDB(sqldb, sqlitedialect.New())
    ctx := context.Background()

    // Create table and insert sample data
    db.NewCreateTable().Model((*User)(nil)).Exec(ctx)
    // Insert sample data...

    // Create a filter for users with age > 20
    filterJSON := `{
        "logic": "and",
        "filters": [
            {"field": "age", "operator": "gt", "value": 20}
        ]
    }`

    // Create a sort by last_name
    sortJSON := `[{"field": "last_name", "dir": "asc"}]`

    // Parse the filter and sort JSON
    ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 10)
    if err != nil {
        panic(err)
    }

    // Create a base query
    query := db.NewSelect().Model((*User)(nil))

    // Apply the BunQL filters, sort, and pagination
    query = ql.Apply(ctx, query)

    // Execute the query
    var users []User
    err = query.Scan(ctx, &users)
    if err != nil {
        panic(err)
    }

    // Print results
    for _, user := range users {
        fmt.Printf("%s %s (Age: %d)\n", user.FirstName, user.LastName, user.Age)
    }
}
```

## Filter JSON Format

Filters are defined using a JSON structure:

```json
{
    "logic": "and",
    "filters": [
        {"field": "age", "operator": "gt", "value": 20}
    ],
    "groups": [
        {
            "logic": "or",
            "filters": [
                {"field": "first_name", "operator": "like", "value": "J%"},
                {"field": "age", "operator": "gt", "value": 40}
            ]
        }
    ]
}
```

### Supported Operators

BunQL supports the following operators for filtering:

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal to | `{"field": "age", "operator": "eq", "value": 30}` |
| `neq` | Not equal to | `{"field": "age", "operator": "neq", "value": 30}` |
| `gt` | Greater than | `{"field": "age", "operator": "gt", "value": 20}` |
| `gte` | Greater than or equal to | `{"field": "age", "operator": "gte", "value": 21}` |
| `lt` | Less than | `{"field": "age", "operator": "lt", "value": 50}` |
| `lte` | Less than or equal to | `{"field": "age", "operator": "lte", "value": 49}` |
| `like` | SQL LIKE operator | `{"field": "first_name", "operator": "like", "value": "J%"}` |
| `in` | In a list of values | `{"field": "age", "operator": "in", "value": [20, 30, 40]}` |
| `notin` | Not in a list of values | `{"field": "age", "operator": "notin", "value": [20, 30, 40]}` |
| `isnull` | Is NULL | `{"field": "email", "operator": "isnull", "value": null}` |
| `isnotnull` | Is NOT NULL | `{"field": "email", "operator": "isnotnull", "value": null}` |
| `between` | Between two values | `{"field": "age", "operator": "between", "value": [20, 30]}` |

For the `between` operator, the `value` must be an array with exactly two elements: the lower and upper bounds (inclusive).

## Sort JSON Format

Sorting is defined using a JSON array:

```json
[
    {"field": "age", "dir": "desc"},
    {"field": "last_name", "dir": "asc"}
]
```

## Field Validation

You can validate that only allowed fields are used for filtering and sorting:

```go
// Define allowed fields for filtering and sorting
allowedFilterFields := []string{"age", "first_name", "last_name"}
allowedSortFields := []string{"age", "last_name"}

// Parse the filter and sort JSON with allowed fields
ql, err := bunql.ParseFromParamsWithAllowedFields(filterJSON, sortJSON, 1, 10, allowedFilterFields, allowedSortFields)
if err != nil {
    // Handle validation error
    panic(err)
}
```

If a filter or sort field is not in the allowed fields list, an error will be returned:
- For filters: `filter field 'email' is not allowed`
- For sorts: `sort field 'email' is not allowed`

## Getting Total Count

You can get the total count of records alongside paginated results:

```go
// Parse the filter and sort JSON with pagination
ql, err := bunql.ParseFromParams(filterJSON, sortJSON, 1, 10)
if err != nil {
    panic(err)
}

// Create a base query
query := db.NewSelect().Model((*User)(nil))

// Apply the BunQL filters, sort, and pagination, and get a count query
mainQuery, countQuery := ql.ApplyWithCount(ctx, query)

// Execute both queries and get the results along with the total count
users, totalCount, err := bunql.ExecuteWithCount[User](ctx, mainQuery, countQuery)
if err != nil {
    panic(err)
}

// Calculate pagination metadata with base URI for prev/next links
metadata := bunql.GetPaginationMetadata(ql.Pagination, totalCount, "https://api.example.com/users")
fmt.Printf("Pagination metadata: %+v\n", metadata)
```

The `metadata` will contain:
- `page`: Current page number
- `pageSize`: Number of items per page
- `total`: Total number of pages
- `totalItems`: Total number of items matching the filter criteria
- `prev`: URL for the previous page (only included if not on the first page)
- `next`: URL for the next page (only included if not on the last page)

Note that any existing query parameters in the base URI will be preserved in the prev and next URLs.

## Testing

The project uses Go's standard testing package along with the testify library for assertions.

To run all tests:
```bash
go test ./...
```

To run specific tests:
```bash
go test -v ./e2e -run TestSimpleQuery
```

To enable SQL query debugging, set the BUNDEBUG environment variable:
```bash
BUNDEBUG=1 go test ./e2e
```

## Project Structure

- `bunql.go`: Main package file with the BunQL struct and methods
- `dto/`: Data transfer objects used for filters, sorting, and pagination
- `filter/`: Filter parsing and application logic
- `sorting/`: Sort parsing and application logic
- `pagination/`: Pagination logic
- `operator/`: SQL operator handling
- `e2e/`: End-to-end tests

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
