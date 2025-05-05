package e2e

import "github.com/uptrace/bun"

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64  `bun:"id,pk,autoincrement"`
	FirstName string `bun:"first_name"`
	LastName  string `bun:"last_name"`
	Email     string `bun:"email"`
	Age       int    `bun:"age"`
	Active    bool   `bun:"active"`
}
