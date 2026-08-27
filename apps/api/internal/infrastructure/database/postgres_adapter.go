package database

import (
	"context"
	"fmt"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter implements the IDBAdapter interface for PostgreSQL
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter creates a new uninitialized PostgreSQL adapter instance
func NewPostgresAdapter() ports.IDBAdapter {
	return &PostgresAdapter{}
}

// Connect initializes the connection pool to the PostgreSQL database
func (p *PostgresAdapter) Connect(ctx context.Context, config domain.DBConnectionConfig) error {
	// Construct the Data Source Name (DSN)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", 
		config.User, config.Password, config.Host, config.Port, config.Database)
	
	// Create connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	
	p.pool = pool
	return p.Ping(ctx)
}

// Disconnect gracefully closes all connections in the pool
func (p *PostgresAdapter) Disconnect() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

// Ping verifies the connection is still alive
func (p *PostgresAdapter) Ping(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("connection pool not initialized")
	}
	return p.pool.Ping(ctx)
}

// GetDatabases retrieves all accessible databases on the server
func (p *PostgresAdapter) GetDatabases(ctx context.Context) ([]string, error) {
	query := "SELECT datname FROM pg_database WHERE datistemplate = false"
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}
	return databases, nil
}

func formatRows(count int64) string {
	if count >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(count)/1000000000.0)
	}
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(count)/1000000.0)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000.0)
	}
	return fmt.Sprintf("%d", count)
}

// GetTables retrieves all tables within the current public schema
func (p *PostgresAdapter) GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error) {
	query := `
		SELECT 
			c.relname as table_name,
			CASE WHEN c.relkind = 'v' THEN 'VIEW' ELSE 'BASE TABLE' END as table_type,
			pg_size_pretty(pg_total_relation_size(c.oid)) as total_size
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' 
		  AND c.relkind IN ('r', 'p', 'v')
	`
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]domain.TableMetadata, 0)
	for rows.Next() {
		var meta domain.TableMetadata
		if err := rows.Scan(&meta.Name, &meta.Type, &meta.Size); err != nil {
			return nil, err
		}
		tables = append(tables, meta)
	}
	
	// Fetch exact row counts securely for each table
	for i, t := range tables {
		sanitizedTable := pgx.Identifier{t.Name}.Sanitize()
		var count int64
		
		// Note: for massive databases with billions of rows, this can be slow,
		// but it guarantees exact counts which is critical for smaller data management.
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", sanitizedTable)
		if err := p.pool.QueryRow(ctx, countQuery).Scan(&count); err != nil {
			count = 0 // Default to 0 if count fails (e.g. lack of permissions)
		}
		
		tables[i].Rows = formatRows(count)
	}

	return tables, nil
}

// GetTableSchema retrieves column metadata for a specific table
func (p *PostgresAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	query := `
		SELECT 
			c.column_name, 
			c.data_type, 
			c.is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tco
			JOIN information_schema.key_column_usage kcu 
			  ON kcu.constraint_name = tco.constraint_name
			  AND kcu.constraint_schema = tco.constraint_schema
			WHERE tco.constraint_type = 'PRIMARY KEY' 
			  AND kcu.table_name = $1
			  AND kcu.table_schema = 'public'
		) pk ON c.column_name = pk.column_name
		WHERE c.table_schema = 'public' AND c.table_name = $1
	`
	rows, err := p.pool.Query(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []domain.ColumnMetadata
	for rows.Next() {
		var meta domain.ColumnMetadata
		var isNullable string
		if err := rows.Scan(&meta.Name, &meta.Type, &isNullable, &meta.IsPrimaryKey); err != nil {
			return nil, err
		}
		meta.IsNullable = (isNullable == "YES")
		columns = append(columns, meta)
	}
	return columns, nil
}

// FetchRows executes a structured SelectQuery securely
func (p *PostgresAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	// A robust implementation would use a SQL builder (e.g., squirrel) to dynamically build this safely.
	// We'll use a hardcoded safe parameterized query for demonstration.
	sanitizedTable := pgx.Identifier{query.Table}.Sanitize()
	sqlStr := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", sanitizedTable)
	return p.ExecuteRaw(ctx, sqlStr, query.Limit, query.Offset)
}

// UpdateRow executes a structured UpdateQuery safely
func (p *PostgresAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	// Placeholder for dynamic update query builder logic
	return 0, fmt.Errorf("UpdateRow not fully implemented yet")
}

// ExecuteRaw runs a raw SQL query. It isolates execution and returns a dynamic ResultSet.
func (p *PostgresAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	rows, err := p.pool.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Extract column names dynamically from the result
	fieldDescriptions := rows.FieldDescriptions()
	var columns []string
	for _, fd := range fieldDescriptions {
		columns = append(columns, string(fd.Name))
	}

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		
		// Map values to their respective column names
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		results = append(results, rowMap)
	}

	return &domain.ResultSet{
		Columns: columns,
		Rows:    results,
	}, nil
}
