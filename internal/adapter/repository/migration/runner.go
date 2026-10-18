package migration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"new-inventory/internal/adapter/repository/dialect"
)

//go:embed scripts/*
var migrationScripts embed.FS

type Runner struct {
	db      *sql.DB
	driver  dialect.Driver
	dialect dialect.Dialect
}

func NewRunner(db *sql.DB, driver dialect.Driver) (*Runner, error) {
	d, err := dialect.Get(driver)
	if err != nil {
		return nil, fmt.Errorf("migration runner error: %w", err)
	}

	return &Runner{
		db:      db,
		driver:  driver,
		dialect: d,
	}, nil
}

// Run executes all pending migrations for the configured database dialect.
func (r *Runner) Run(ctx context.Context) ([]string, error) {
	if err := r.ensureMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	appliedMap, err := r.getAppliedMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve applied migrations: %w", err)
	}

	dirPath := path.Join("scripts", string(r.driver))
	entries, err := fs.ReadDir(migrationScripts, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory '%s': %w", dirPath, err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	var newlyApplied []string
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, ".up.sql")
		if appliedMap[version] {
			continue
		}

		filePath := path.Join(dirPath, filename)
		content, err := migrationScripts.ReadFile(filePath)
		if err != nil {
			return newlyApplied, fmt.Errorf("failed to read migration file '%s': %w", filePath, err)
		}

		if err := r.applyMigration(ctx, version, string(content)); err != nil {
			return newlyApplied, fmt.Errorf("failed to apply migration '%s': %w", version, err)
		}

		newlyApplied = append(newlyApplied, version)
	}

	return newlyApplied, nil
}

// AppliedVersions returns all migration versions currently applied to the database.
func (r *Runner) AppliedVersions(ctx context.Context) ([]string, error) {
	if err := r.ensureMigrationsTable(ctx); err != nil {
		return nil, err
	}

	query := "SELECT version FROM schema_migrations ORDER BY version ASC"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *Runner) ensureMigrationsTable(ctx context.Context) error {
	var createTableSQL string
	switch r.driver {
	case dialect.DriverPostgres:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version VARCHAR(64) PRIMARY KEY,
				applied_at TIMESTAMPTZ NOT NULL
			);
		`
	case dialect.DriverOracle:
		createTableSQL = `
			BEGIN
				EXECUTE IMMEDIATE 'CREATE TABLE schema_migrations (version VARCHAR2(64) PRIMARY KEY, applied_at TIMESTAMP WITH TIME ZONE NOT NULL)';
			EXCEPTION
				WHEN OTHERS THEN
					IF SQLCODE != -955 THEN
						RAISE;
					END IF;
			END;
		`
	default:
		// SQLite and generic ANSI SQL
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version VARCHAR(64) PRIMARY KEY,
				applied_at TEXT NOT NULL
			);
		`
	}

	_, err := r.db.ExecContext(ctx, createTableSQL)
	return err
}

func (r *Runner) getAppliedMap(ctx context.Context) (map[string]bool, error) {
	applied, err := r.AppliedVersions(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(applied))
	for _, v := range applied {
		m[v] = true
	}
	return m, nil
}

func (r *Runner) applyMigration(ctx context.Context, version, script string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	statements := splitSQLStatements(script)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("statement error: %w\nSQL: %s", err, stmt)
		}
	}

	insertVersionQuery := r.dialect.Rebind("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)")
	now := r.dialect.FormatTime(time.Now().UTC())
	if _, err := tx.ExecContext(ctx, insertVersionQuery, version, now); err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	return tx.Commit()
}

func splitSQLStatements(script string) []string {
	var statements []string
	var sb strings.Builder

	inSingleQuote := false
	runes := []rune(script)
	n := len(runes)

	for i := 0; i < n; i++ {
		ch := runes[i]
		if ch == '\'' {
			if inSingleQuote && i+1 < n && runes[i+1] == '\'' {
				sb.WriteRune(ch)
				sb.WriteRune('\'')
				i++
				continue
			}
			inSingleQuote = !inSingleQuote
			sb.WriteRune(ch)
			continue
		}

		if ch == ';' && !inSingleQuote {
			stmt := strings.TrimSpace(sb.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			sb.Reset()
		} else {
			sb.WriteRune(ch)
		}
	}

	leftover := strings.TrimSpace(sb.String())
	if leftover != "" {
		statements = append(statements, leftover)
	}

	return statements
}
