package database

import (
	"context"
	"database/sql"
	"fmt"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLAdapter struct {
	db *sql.DB
}

func NewMySQLAdapter() ports.IDBAdapter {
	return &MySQLAdapter{}
}

func init() {
	RegisterAdapter("mysql", NewMySQLAdapter)
}

func (m *MySQLAdapter) Connect(ctx context.Context, config domain.DBConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", 
		config.User, config.Password, config.Host, config.Port, config.Database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}
	
	m.db = db
	return m.Ping(ctx)
}

func (m *MySQLAdapter) Disconnect() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *MySQLAdapter) Ping(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("mysql connection not initialized")
	}
	return m.db.PingContext(ctx)
}

func (m *MySQLAdapter) GetDatabases(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, "SHOW DATABASES")
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

func (m *MySQLAdapter) GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error) {
	query := "SELECT TABLE_NAME, TABLE_TYPE, TABLE_ROWS, (DATA_LENGTH + INDEX_LENGTH) AS SIZE FROM information_schema.tables WHERE TABLE_SCHEMA = ?"
	rows, err := m.db.QueryContext(ctx, query, database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []domain.TableMetadata
	for rows.Next() {
		var meta domain.TableMetadata
		var rowCount sql.NullInt64
		var size sql.NullInt64
		if err := rows.Scan(&meta.Name, &meta.Type, &rowCount, &size); err != nil {
			return nil, err
		}
		if rowCount.Valid {
			meta.Rows = fmt.Sprintf("%d", rowCount.Int64)
		} else {
			meta.Rows = "0"
		}
		if size.Valid {
			val := size.Int64
			if val < 1024 {
				meta.Size = fmt.Sprintf("%d B", val)
			} else if val < 1024*1024 {
				meta.Size = fmt.Sprintf("%d KB", val/1024)
			} else {
				meta.Size = fmt.Sprintf("%d MB", val/(1024*1024))
			}
		} else {
			meta.Size = "Unknown"
		}
		tables = append(tables, meta)
	}
	return tables, nil
}

func (m *MySQLAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	query := `
		SELECT 
			COLUMN_NAME, 
			COLUMN_TYPE, 
			IS_NULLABLE, 
			COLUMN_KEY
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ORDINAL_POSITION
	`
	rows, err := m.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []domain.ColumnMetadata
	for rows.Next() {
		var col domain.ColumnMetadata
		var isNullable, colKey string
		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &colKey); err != nil {
			return nil, err
		}
		
		col.IsNullable = isNullable == "YES"
		col.IsPrimaryKey = colKey == "PRI"
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (m *MySQLAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	qStr := fmt.Sprintf("SELECT * FROM `%s`", query.Table)
	if query.Limit > 0 {
		qStr += fmt.Sprintf(" LIMIT %d OFFSET %d", query.Limit, query.Offset)
	}
	return m.ExecuteRaw(ctx, qStr)
}

func (m *MySQLAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	if len(update.Updates) == 0 {
		return 0, fmt.Errorf("no columns to update")
	}
	
	setClauses := ""
	var args []interface{}
	for k, v := range update.Updates {
		if setClauses != "" {
			setClauses += ", "
		}
		setClauses += fmt.Sprintf("`%s` = ?", k)
		args = append(args, v)
	}
	
	whereClauses := ""
	for k, v := range update.Conditions {
		if whereClauses != "" {
			whereClauses += " AND "
		}
		whereClauses += fmt.Sprintf("`%s` = ?", k)
		args = append(args, v)
	}
	
	qStr := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", update.Table, setClauses, whereClauses)
	
	res, err := m.db.ExecContext(ctx, qStr, args...)
	if err != nil {
		return 0, err
	}
	
	return res.RowsAffected()
}

func (m *MySQLAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	rows, err := m.db.QueryContext(ctx, query, params...)
	if err != nil {
		// Fallback for DML queries that might not return rows in some driver configurations
		res, execErr := m.db.ExecContext(ctx, query, params...)
		if execErr != nil {
			return nil, err // Return original error
		}
		affected, _ := res.RowsAffected()
		return &domain.ResultSet{
			Columns: []string{"rows_affected"},
			Rows:    []map[string]interface{}{{"rows_affected": affected}},
		}, nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if val == nil {
				rowMap[colName] = nil
				continue
			}
			if b, ok := (*val).([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = *val
			}
		}
		results = append(results, rowMap)
	}

	return &domain.ResultSet{
		Columns: cols,
		Rows:    results,
	}, nil
}
