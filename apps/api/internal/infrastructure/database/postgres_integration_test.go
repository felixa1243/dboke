package database

import (
	"context"
	"os"
	"strconv"
	"testing"

	"dboke-api/internal/core/domain"
)

// Runs only when DBOKE_TEST_PG_HOST is set, e.g.
//   docker run --rm -d -p 15432:5432 -e POSTGRES_PASSWORD=pw postgres:16-alpine
//   DBOKE_TEST_PG_HOST=localhost DBOKE_TEST_PG_PORT=15432 DBOKE_TEST_PG_PASSWORD=pw go test ./internal/infrastructure/database
func newTestPostgres(t *testing.T) *PostgresAdapter {
	t.Helper()
	host := os.Getenv("DBOKE_TEST_PG_HOST")
	if host == "" {
		t.Skip("DBOKE_TEST_PG_HOST not set")
	}
	port, _ := strconv.Atoi(os.Getenv("DBOKE_TEST_PG_PORT"))
	a := &PostgresAdapter{}
	err := a.Connect(context.Background(), domain.DBConnectionConfig{
		Host: host, Port: port,
		User:     "postgres",
		Password: os.Getenv("DBOKE_TEST_PG_PASSWORD"),
		Database: "postgres",
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { a.Disconnect() })
	return a
}

func TestPostgres_ExecuteRaw_Results(t *testing.T) {
	a := newTestPostgres(t)
	ctx := context.Background()

	if _, err := a.ExecuteRaw(ctx, "CREATE TEMP TABLE dboke_t (id int PRIMARY KEY, v text)"); err != nil {
		t.Fatal(err)
	}

	// DML without a result set reports rows_affected
	rs, err := a.ExecuteRaw(ctx, "INSERT INTO dboke_t VALUES (1,'a'),(2,'b'),(3,'c')")
	if err != nil {
		t.Fatal(err)
	}
	if len(rs.Columns) != 1 || rs.Columns[0] != "rows_affected" || rs.Rows[0]["rows_affected"] != int64(3) {
		t.Errorf("INSERT result = %+v", rs)
	}

	// A SELECT returning nothing has empty (non-nil) columns' rows, not null
	rs, err = a.ExecuteRaw(ctx, "SELECT * FROM dboke_t WHERE id > 100")
	if err != nil {
		t.Fatal(err)
	}
	if rs.Rows == nil || len(rs.Rows) != 0 || len(rs.Columns) != 2 {
		t.Errorf("empty SELECT result = %+v (rows nil: %v)", rs, rs.Rows == nil)
	}

	// Constraint violations surface as errors instead of an empty success
	if _, err := a.ExecuteRaw(ctx, "INSERT INTO dboke_t VALUES (1,'dup')"); err == nil {
		t.Error("duplicate key insert returned no error")
	}
}
