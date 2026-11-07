package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config encapsulates connection details for a specific database server.
type Config struct {
	Driver          string        // "sqlite", "mysql", "postgres", "sqlserver"
	Host            string        // Server host (e.g. 127.0.0.1)
	Port            int           // Server port
	Database        string        // Database name or file path
	Username        string        // User login
	Password        string        // User password
	Charset         string        // Character set (default utf8mb4)
	SSLMode         string        // SSL mode for PostgreSQL (e.g. disable, prefer)
	DSN             string        // Optional custom DSN string
	MaxOpenConns    int           // Max open connection pool limit
	MaxIdleConns    int           // Max idle connections
	ConnMaxLifetime time.Duration // Connection lifetime duration
}

// FormatDSN generates the standard DSN and standard driver name for database/sql.
func (c Config) FormatDSN() (string, string) {
	if c.DSN != "" {
		return c.Driver, c.DSN
	}

	driver := strings.ToLower(c.Driver)
	switch driver {
	case "mysql", "mariadb":
		charset := c.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		host := c.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := c.Port
		if port <= 0 {
			port = 3306
		}
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, addr, c.Database, charset)
		return "mysql", dsn

	case "postgres", "postgresql", "pgsql":
		host := c.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := c.Port
		if port <= 0 {
			port = 5432
		}
		sslmode := c.SSLMode
		if sslmode == "" {
			sslmode = "prefer"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, c.Username, c.Password, c.Database, sslmode)
		return "postgres", dsn

	case "sqlserver", "mssql":
		host := c.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := c.Port
		if port <= 0 {
			port = 1433
		}
		dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			host, c.Username, c.Password, port, c.Database)
		return "sqlserver", dsn

	default: // sqlite
		dbPath := c.Database
		if dbPath == "" {
			dbPath = "database.sqlite"
		}
		return "sqlite", dbPath
	}
}

// Connection represents an active database connection configured with a specific Dialect.
type Connection struct {
	name    string
	config  Config
	dialect Dialect
	db      *sql.DB
}

// Name returns the identifier of this connection.
func (c *Connection) Name() string {
	return c.name
}

// Config returns the configuration settings.
func (c *Connection) Config() Config {
	return c.config
}

// Dialect returns the active SQL dialect.
func (c *Connection) Dialect() Dialect {
	return c.dialect
}

// DB returns the underlying standard library *sql.DB pool.
func (c *Connection) DB() *sql.DB {
	return c.db
}

// Table returns a new QueryBuilder linked directly to this connection and its dialect.
func (c *Connection) Table(tableName string) *QueryBuilder {
	qb := Table(tableName)
	qb.SetDialect(c.dialect)
	if c.db != nil {
		qb.SetExecutor(c.db)
	}
	return qb
}

// Execute runs an arbitrary query without returning rows.
func (c *Connection) Execute(query string, args ...any) (sql.Result, error) {
	if c.db == nil {
		return nil, errors.New("database: connection is not open")
	}
	return c.db.Exec(query, args...)
}

// Query runs a query that returns rows.
func (c *Connection) Query(query string, args ...any) (*sql.Rows, error) {
	if c.db == nil {
		return nil, errors.New("database: connection is not open")
	}
	return c.db.Query(query, args...)
}

// QueryRow runs a query that returns at most one row.
func (c *Connection) QueryRow(query string, args ...any) *sql.Row {
	if c.db == nil {
		return nil
	}
	return c.db.QueryRow(query, args...)
}

// Transaction executes a function within an isolated database transaction.
func (c *Connection) Transaction(fn func(tx *sql.Tx) error) error {
	if c.db == nil {
		return errors.New("database: connection is not open")
	}
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Close closes the underlying database pool.
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------
// Multi-Database Connection Manager
// -----------------------------------------------------------------------------

// Manager maintains multiple named database connections and configurations.
type Manager struct {
	mu          sync.RWMutex
	defaultName string
	configs     map[string]Config
	connections map[string]*Connection
}

// NewManager initializes an empty Multi-Database Manager.
func NewManager() *Manager {
	return &Manager{
		defaultName: "default",
		configs:     make(map[string]Config),
		connections: make(map[string]*Connection),
	}
}

// AddConnection registers a database configuration under a given connection name.
func (m *Manager) AddConnection(name string, cfg Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[name] = cfg
}

// SetDefault sets the default connection identifier.
func (m *Manager) SetDefault(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultName = name
}

// Connection retrieves an active connection by name (or default if omitted).
// If the connection pool has not yet been opened, it initializes it lazily.
func (m *Manager) Connection(name ...string) (*Connection, error) {
	targetName := m.defaultName
	if len(name) > 0 && name[0] != "" {
		targetName = name[0]
	}

	m.mu.RLock()
	conn, exists := m.connections[targetName]
	cfg, cfgExists := m.configs[targetName]
	m.mu.RUnlock()

	if exists && conn != nil {
		return conn, nil
	}

	if !cfgExists {
		// Fallback to environment default if requested default is not configured
		if targetName == "default" {
			cfg = configFromEnv()
		} else {
			return nil, fmt.Errorf("database: connection '%s' is not configured", targetName)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check after lock
	if conn, exists := m.connections[targetName]; exists && conn != nil {
		return conn, nil
	}

	dialect := GetDialect(cfg.Driver)

	// Attempt to open sql.DB if driver is available
	var db *sql.DB
	driverName, dsn := cfg.FormatDSN()

	// Check if driver is registered in database/sql
	hasDriver := false
	for _, registered := range sql.Drivers() {
		if registered == driverName || (driverName == "sqlite" && (registered == "sqlite3" || registered == "modernc.org/sqlite")) {
			hasDriver = true
			if driverName == "sqlite" && registered != "sqlite" {
				driverName = registered
			}
			break
		}
	}

	if hasDriver {
		openedDB, err := sql.Open(driverName, dsn)
		if err == nil {
			if cfg.MaxOpenConns > 0 {
				openedDB.SetMaxOpenConns(cfg.MaxOpenConns)
			}
			if cfg.MaxIdleConns > 0 {
				openedDB.SetMaxIdleConns(cfg.MaxIdleConns)
			}
			if cfg.ConnMaxLifetime > 0 {
				openedDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
			}
			db = openedDB
		}
	}

	newConn := &Connection{
		name:    targetName,
		config:  cfg,
		dialect: dialect,
		db:      db,
	}

	m.connections[targetName] = newConn
	return newConn, nil
}

// CloseAll closes all active database connections.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for name, conn := range m.connections {
		if conn.db != nil {
			if err := conn.db.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		delete(m.connections, name)
	}
	return firstErr
}

// -----------------------------------------------------------------------------
// Global Instance & Convenience Shortcuts
// -----------------------------------------------------------------------------

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

// GetManager returns the singleton Multi-Database Manager instance.
func GetManager() *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewManager()
		defaultManager.AddConnection("default", configFromEnv())
	})
	return defaultManager
}

// Connection retrieves a database connection from the global manager.
// Example: database.Connection("mysql") or database.Connection() for default.
func GetConnection(name ...string) (*Connection, error) {
	return GetManager().Connection(name...)
}

func configFromEnv() Config {
	driver := os.Getenv("DB_CONNECTION")
	if driver == "" {
		driver = "sqlite"
	}
	host := os.Getenv("DB_HOST")
	port := 0
	if pStr := os.Getenv("DB_PORT"); pStr != "" {
		port, _ = strconv.Atoi(pStr)
	}
	database := os.Getenv("DB_DATABASE")
	if database == "" && driver == "sqlite" {
		database = "database.sqlite"
	}
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")

	return Config{
		Driver:   driver,
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
	}
}

