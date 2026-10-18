package migration_test

import (
	"context"
	"database/sql"
	"testing"

	"new-inventory/internal/adapter/repository/dialect"
	"new-inventory/internal/adapter/repository/migration"

	_ "modernc.org/sqlite"
)

func TestMigrationRunner_SQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	runner, err := migration.NewRunner(db, dialect.DriverSQLite)
	if err != nil {
		t.Fatalf("failed to create migration runner: %v", err)
	}

	ctx := context.Background()

	// 1. Initial run: should apply 000001_init_schema
	applied, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if len(applied) == 0 {
		t.Fatalf("expected at least 1 migration applied, got %d", len(applied))
	}
	if applied[0] != "000001_init_schema" {
		t.Errorf("expected applied migration '000001_init_schema', got '%s'", applied[0])
	}

	// 2. Verify schema_migrations table contains the version
	versions, err := runner.AppliedVersions(ctx)
	if err != nil {
		t.Fatalf("failed to get applied versions: %v", err)
	}
	if len(versions) != 1 || versions[0] != "000001_init_schema" {
		t.Errorf("expected versions ['000001_init_schema'], got %v", versions)
	}

	// 3. Verify all tables exist and can be queried
	tables := []string{
		"items",
		"stock_levels",
		"stock_movements",
		"stock_cost_layers",
		"tenant_settings",
		"suppliers",
		"purchase_orders",
		"purchase_order_items",
	}
	for _, tbl := range tables {
		var name string
		row := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl)
		if err := row.Scan(&name); err != nil {
			t.Errorf("table '%s' does not exist in database: %v", tbl, err)
		}
	}

	// 4. Idempotency: running again should apply 0 migrations
	appliedSecond, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(appliedSecond) != 0 {
		t.Errorf("expected 0 migrations on second run, got %d", len(appliedSecond))
	}
}

func TestMigrationRunner_UnsupportedDriver(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = migration.NewRunner(db, "unsupported")
	if err == nil {
		t.Error("expected error for unsupported driver, got nil")
	}
}
