package domain

// DBConnectionConfig holds the necessary information to connect to a target database
type DBConnectionConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// TableMetadata represents basic information about a database table
type TableMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"` // e.g., "BASE TABLE", "VIEW"
	Rows string `json:"rows"`
	Size string `json:"size"`
}

// ColumnMetadata holds information about a specific column in a table
type ColumnMetadata struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimaryKey bool `json:"is_primary_key"`
}

// SelectQuery represents a structured, safe query to fetch rows
type SelectQuery struct {
	Table   string
	Columns []string
	Limit   int
	Offset  int
	// Add Where clause abstractions here later
}

// UpdateQuery represents a structured, safe update
type UpdateQuery struct {
	Table      string
	Updates    map[string]interface{}
	Conditions map[string]interface{}
}

// ResultSet represents the data returned from a query
type ResultSet struct {
	Columns []string                   `json:"columns"`
	Rows    []map[string]interface{}   `json:"rows"`
}
