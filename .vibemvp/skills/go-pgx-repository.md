# Go PostgreSQL Repository (pgx/v5)

## The PgxPool Interface (MANDATORY)
Define in `internal/repository/interfaces.go`. NEVER use `*pgxpool.Pool` in struct fields —
doing so prevents pgxmock injection in tests and causes the type mismatch compile error.

```go
import (
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

type PgxPool interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
    Begin(ctx context.Context) (pgx.Tx, error)
}
```

## Repository Struct Pattern
```go
type userRepository struct {
    pool PgxPool  // interface, NOT *pgxpool.Pool
}

func NewUserRepository(pool PgxPool) *userRepository {
    return &userRepository{pool: pool}
}
```

## Test Injection Pattern (pgxmock)
```go
func TestUserRepository_Create(t *testing.T) {
    pool, err := pgxmock.NewPool()
    if err != nil {
        t.Fatal(err)
    }
    repo := NewUserRepository(pool)  // inject via PgxPool interface — no cast
    pool.ExpectQuery(`INSERT INTO users`).WillReturnRows(...)
}
```

## go.mod — Direct Dependencies Only
List only what your code directly imports. The dependency resolution step runs
`go mod tidy` to add transitive deps from the module proxy automatically.
Do NOT guess pseudo-versions for transitive packages — this is handled for you.

```
require (
    github.com/jackc/pgx/v5 v5.5.5
    github.com/pashagolub/pgxmock/v3 v3.3.0
)
```

## Connection Setup (production)
```go
func NewPool(ctx context.Context, dsn string) (PgxPool, error) {
    return pgxpool.New(ctx, dsn)
}
```
`*pgxpool.Pool` satisfies `PgxPool` automatically — no adapter required.
