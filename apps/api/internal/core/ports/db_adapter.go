package ports

import (
	"context"
	"dboke-api/internal/core/domain"
)

// IDBAdapter defines the contract for interacting with any supported database (e.g., Postgres, MySQL)
type IDBAdapter interface {
	// Connection Management
	Connect(ctx context.Context, config domain.DBConnectionConfig) error
	Disconnect() error
	Ping(ctx context.Context) error

	// Schema & Metadata Exploration
	GetDatabases(ctx context.Context) ([]string, error)
	GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error)
	GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error)

	// Data Operations (Safe, parameterized operations for UI components)
	FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error)
	UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error)
	
	// Raw Query Execution
	// Must be isolated. Backend will enforce statement timeouts and read-only flags based on RBAC.
	ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error)
}
