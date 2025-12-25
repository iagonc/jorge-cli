package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// DatabaseUsecase handles database operations
type DatabaseUsecase struct {
	logger *zap.Logger
}

// NewDatabaseUsecase creates a new DatabaseUsecase
func NewDatabaseUsecase(logger *zap.Logger) *DatabaseUsecase {
	return &DatabaseUsecase{logger: logger}
}

// Connect tests database connection
func (u *DatabaseUsecase) Connect(ctx context.Context, dsn, dbType string) (*models.DatabaseInfo, error) {
	info := &models.DatabaseInfo{
		Type:      dbType,
		Connected: false,
	}

	// Parse DSN for display
	info.Host, info.Port, info.Database = u.parseDSN(dsn, dbType)

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		info.Error = err.Error()
		return info, nil
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)

	// Ping database
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctxTimeout); err != nil {
		info.Error = err.Error()
		return info, nil
	}

	info.Connected = true

	// Get version and other info based on database type
	switch dbType {
	case "postgres":
		u.getPostgresInfo(ctx, db, info)
	case "mysql":
		u.getMySQLInfo(ctx, db, info)
	}

	return info, nil
}

func (u *DatabaseUsecase) parseDSN(dsn, dbType string) (host string, port int, database string) {
	// Simplified parsing
	switch dbType {
	case "postgres":
		// postgres://user:pass@host:port/db
		if strings.Contains(dsn, "@") {
			parts := strings.Split(dsn, "@")
			if len(parts) > 1 {
				hostPart := strings.Split(parts[1], "/")
				if len(hostPart) > 0 {
					hostPort := strings.Split(hostPart[0], ":")
					host = hostPort[0]
					if len(hostPort) > 1 {
						fmt.Sscanf(hostPort[1], "%d", &port)
					} else {
						port = 5432
					}
				}
				if len(hostPart) > 1 {
					database = strings.Split(hostPart[1], "?")[0]
				}
			}
		}
	case "mysql":
		// user:pass@tcp(host:port)/db
		if strings.Contains(dsn, "@tcp(") {
			start := strings.Index(dsn, "@tcp(") + 5
			end := strings.Index(dsn[start:], ")")
			hostPort := dsn[start : start+end]
			parts := strings.Split(hostPort, ":")
			host = parts[0]
			if len(parts) > 1 {
				fmt.Sscanf(parts[1], "%d", &port)
			} else {
				port = 3306
			}
			if idx := strings.Index(dsn, ")/"); idx != -1 {
				database = strings.Split(dsn[idx+2:], "?")[0]
			}
		}
	}
	return
}

func (u *DatabaseUsecase) getPostgresInfo(ctx context.Context, db *sql.DB, info *models.DatabaseInfo) {
	// Version
	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err == nil {
		info.Version = version
	}

	// Connection count
	var activeConns, maxConns int
	db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConns)
	db.QueryRowContext(ctx, "SHOW max_connections").Scan(&maxConns)
	info.ActiveConns = activeConns
	info.MaxConns = maxConns

	// Database size
	var dbSize string
	if err := db.QueryRowContext(ctx, "SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize); err == nil {
		info.DatabaseSize = dbSize
	}

	// Table count
	var tableCount int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount); err == nil {
		info.TableCount = tableCount
	}
}

func (u *DatabaseUsecase) getMySQLInfo(ctx context.Context, db *sql.DB, info *models.DatabaseInfo) {
	// Version
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err == nil {
		info.Version = version
	}

	// Connection count
	var activeConns int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.processlist").Scan(&activeConns)
	info.ActiveConns = activeConns

	// Max connections
	var maxConns int
	db.QueryRowContext(ctx, "SHOW VARIABLES LIKE 'max_connections'").Scan(new(string), &maxConns)
	info.MaxConns = maxConns

	// Database size (if database is specified)
	if info.Database != "" {
		var dbSize string
		query := fmt.Sprintf(`SELECT CONCAT(ROUND(SUM(data_length + index_length) / 1024 / 1024, 2), ' MB')
			FROM information_schema.tables WHERE table_schema = '%s'`, info.Database)
		if err := db.QueryRowContext(ctx, query).Scan(&dbSize); err == nil {
			info.DatabaseSize = dbSize
		}
	}
}

// ListTables lists tables in a database
func (u *DatabaseUsecase) ListTables(ctx context.Context, dsn, dbType string) ([]models.TableInfo, error) {
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var tables []models.TableInfo

	switch dbType {
	case "postgres":
		rows, err := db.QueryContext(ctx, `
			SELECT
				t.table_name,
				t.table_schema,
				COALESCE(s.n_live_tup, 0) as row_count,
				pg_size_pretty(pg_total_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))) as size,
				(SELECT count(*) FROM pg_indexes WHERE tablename = t.table_name) as index_count
			FROM information_schema.tables t
			LEFT JOIN pg_stat_user_tables s ON t.table_name = s.relname
			WHERE t.table_schema = 'public'
			ORDER BY t.table_name`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var t models.TableInfo
			if err := rows.Scan(&t.Name, &t.Schema, &t.RowCount, &t.Size, &t.IndexCount); err == nil {
				tables = append(tables, t)
			}
		}

	case "mysql":
		rows, err := db.QueryContext(ctx, `
			SELECT
				table_name,
				table_schema,
				table_rows,
				CONCAT(ROUND((data_length + index_length) / 1024 / 1024, 2), ' MB') as size
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			ORDER BY table_name`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var t models.TableInfo
			if err := rows.Scan(&t.Name, &t.Schema, &t.RowCount, &t.Size); err == nil {
				tables = append(tables, t)
			}
		}
	}

	return tables, nil
}

// Query executes a read-only query
func (u *DatabaseUsecase) Query(ctx context.Context, dsn, dbType, query string, limit int) (*models.QueryResult, error) {
	result := &models.QueryResult{
		Query: query,
	}

	// Only allow SELECT queries
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(queryUpper, "SELECT") && !strings.HasPrefix(queryUpper, "SHOW") && !strings.HasPrefix(queryUpper, "EXPLAIN") {
		result.Error = "Only SELECT, SHOW, and EXPLAIN queries are allowed"
		return result, nil
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer db.Close()

	// Add LIMIT if not present
	if limit > 0 && !strings.Contains(queryUpper, "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	start := time.Now()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer rows.Close()

	result.Duration = time.Since(start).String()

	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	result.Columns = columns

	// Scan rows
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result.Rows = append(result.Rows, row)
	}

	result.RowCount = len(result.Rows)
	return result, nil
}

// HealthCheck performs a database health check
func (u *DatabaseUsecase) HealthCheck(ctx context.Context, dsn, dbType string) (*models.DatabaseHealth, error) {
	health := &models.DatabaseHealth{
		Type:    dbType,
		Healthy: false,
	}

	// Parse host from DSN
	health.Host, _, _ = u.parseDSN(dsn, dbType)

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		health.Error = err.Error()
		return health, nil
	}
	defer db.Close()

	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		health.Error = err.Error()
		return health, nil
	}
	health.Latency = time.Since(start).String()
	health.Healthy = true

	return health, nil
}
