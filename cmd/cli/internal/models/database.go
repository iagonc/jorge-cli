package models

// DatabaseInfo represents database connection info
type DatabaseInfo struct {
	Type           string `json:"type"` // postgres, mysql, sqlite
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Database       string `json:"database"`
	Connected      bool   `json:"connected"`
	Version        string `json:"version"`
	Uptime         string `json:"uptime,omitempty"`
	ActiveConns    int    `json:"active_connections"`
	MaxConns       int    `json:"max_connections"`
	DatabaseSize   string `json:"database_size"`
	TableCount     int    `json:"table_count"`
	Error          string `json:"error,omitempty"`
}

// TableInfo represents database table information
type TableInfo struct {
	Name       string `json:"name"`
	Schema     string `json:"schema,omitempty"`
	RowCount   int64  `json:"row_count"`
	Size       string `json:"size"`
	IndexCount int    `json:"index_count"`
}

// QueryResult represents a database query result
type QueryResult struct {
	Query       string                   `json:"query"`
	Columns     []string                 `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"row_count"`
	Duration    string                   `json:"duration"`
	Error       string                   `json:"error,omitempty"`
}

// DatabaseHealth represents database health check result
type DatabaseHealth struct {
	Type           string  `json:"type"`
	Host           string  `json:"host"`
	Healthy        bool    `json:"healthy"`
	Latency        string  `json:"latency"`
	ReplicationLag string  `json:"replication_lag,omitempty"`
	DiskUsage      float64 `json:"disk_usage_percent,omitempty"`
	Error          string  `json:"error,omitempty"`
}
