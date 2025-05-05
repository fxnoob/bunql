# BunQL

BunQL is a Go library that provides a structured way to build SQL queries with filtering, sorting, and pagination capabilities for the [Bun SQL toolkit](https://github.com/uptrace/bun).

## Features

- **Dynamic Filtering**: Create complex filter conditions from JSON configurations
- **Flexible Sorting**: Sort by multiple fields with customizable directions
- **Pagination**: Easily paginate results with metadata
- **JSON-based Configuration**: Define filters and sorting using simple JSON structures
- **Fluent API**: Use method chaining for a clean and readable query building experience

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

## Sort JSON Format

Sorting is defined using a JSON array:

```json
[
    {"field": "age", "dir": "desc"},
    {"field": "last_name", "dir": "asc"}
]
```

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