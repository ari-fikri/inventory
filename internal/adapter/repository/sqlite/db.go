package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"new-inventory/internal/adapter/repository/dialect"
	"new-inventory/internal/adapter/repository/migration"
	reposql "new-inventory/internal/adapter/repository/sql"

	_ "modernc.org/sqlite"
)

// OpenDB opens a SQLite database and applies versioned migrations automatically.
func OpenDB(dataSourceName string) (*sql.DB, error) {
	dbWrapper, err := reposql.Open(reposql.Config{
		Driver:        string(dialect.DriverSQLite),
		DSN:           dataSourceName,
		RunMigrations: true,
	})
	if err != nil {
		return nil, err
	}
	return dbWrapper.DB, nil
}

// Migrate executes pending SQLite migrations using the embedded migration runner.
func Migrate(db *sql.DB) error {
	runner, err := migration.NewRunner(db, dialect.DriverSQLite)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}
	_, err = runner.Run(context.Background())
	return err
}
