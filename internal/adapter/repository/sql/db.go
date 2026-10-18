package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"new-inventory/internal/adapter/repository/dialect"
	"new-inventory/internal/adapter/repository/migration"

	_ "modernc.org/sqlite"
)

type Config struct {
	Driver        string
	DSN           string
	RunMigrations bool
}

type DB struct {
	*sql.DB
	Driver  dialect.Driver
	Dialect dialect.Dialect
}

// Open initializes a database connection, discovers the appropriate SQL dialect,
// and optionally applies versioned schema migrations.
func Open(cfg Config) (*DB, error) {
	driverStr := strings.TrimSpace(cfg.Driver)
	if driverStr == "" {
		driverStr = string(dialect.DriverSQLite)
	}

	driver := dialect.Driver(strings.ToLower(driverStr))
	d, err := dialect.Get(driver)
	if err != nil {
		return nil, fmt.Errorf("unsupported database driver: %w", err)
	}

	sqlDriverName := string(driver)
	if driver == dialect.DriverSQLite {
		sqlDriverName = "sqlite"
	}

	db, err := sql.Open(sqlDriverName, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if cfg.RunMigrations {
		runner, err := migration.NewRunner(db, driver)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to create migration runner: %w", err)
		}
		if _, err := runner.Run(context.Background()); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to execute migrations: %w", err)
		}
	}

	return &DB{
		DB:      db,
		Driver:  driver,
		Dialect: d,
	}, nil
}
