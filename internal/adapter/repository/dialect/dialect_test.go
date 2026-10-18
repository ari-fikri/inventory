package dialect_test

import (
	"errors"
	"testing"
	"time"

	"new-inventory/internal/adapter/repository/dialect"
)

func TestDialectRegistry(t *testing.T) {
	drivers := []dialect.Driver{
		dialect.DriverSQLite,
		dialect.DriverPostgres,
		dialect.DriverOracle,
	}

	for _, drv := range drivers {
		d, err := dialect.Get(drv)
		if err != nil {
			t.Fatalf("expected dialect for driver %s, got error: %v", drv, err)
		}
		if d.DriverName() != drv {
			t.Errorf("expected driver name %s, got %s", drv, d.DriverName())
		}
	}

	_, err := dialect.Get("unsupported")
	if err == nil {
		t.Error("expected error for unsupported driver, got nil")
	}
}

func TestSQLiteDialect_Rebind(t *testing.T) {
	d, err := dialect.Get(dialect.DriverSQLite)
	if err != nil {
		t.Fatal(err)
	}

	query := "SELECT id, name FROM items WHERE tenant_id = ? AND sku = ?"
	rebound := d.Rebind(query)
	if rebound != query {
		t.Errorf("expected SQLite rebind to leave '?' unchanged, got: %s", rebound)
	}
}

func TestPostgresDialect_Rebind(t *testing.T) {
	d, err := dialect.Get(dialect.DriverPostgres)
	if err != nil {
		t.Fatal(err)
	}

	query := "SELECT id, name FROM items WHERE tenant_id = ? AND sku = ? AND active = ?"
	expected := "SELECT id, name FROM items WHERE tenant_id = $1 AND sku = $2 AND active = $3"
	rebound := d.Rebind(query)
	if rebound != expected {
		t.Errorf("expected '%s', got '%s'", expected, rebound)
	}

	// Should not rebind ? inside string literals
	queryWithLiteral := "SELECT id, 'why?' AS note FROM items WHERE tenant_id = ? AND sku = ?"
	expectedWithLiteral := "SELECT id, 'why?' AS note FROM items WHERE tenant_id = $1 AND sku = $2"
	reboundWithLiteral := d.Rebind(queryWithLiteral)
	if reboundWithLiteral != expectedWithLiteral {
		t.Errorf("expected '%s', got '%s'", expectedWithLiteral, reboundWithLiteral)
	}
}

func TestOracleDialect_Rebind(t *testing.T) {
	d, err := dialect.Get(dialect.DriverOracle)
	if err != nil {
		t.Fatal(err)
	}

	query := "SELECT id, name FROM items WHERE tenant_id = ? AND sku = ?"
	expected := "SELECT id, name FROM items WHERE tenant_id = :1 AND sku = :2"
	rebound := d.Rebind(query)
	if rebound != expected {
		t.Errorf("expected '%s', got '%s'", expected, rebound)
	}
}

func TestDialect_UniqueConstraintError(t *testing.T) {
	sqliteD, _ := dialect.Get(dialect.DriverSQLite)
	pgD, _ := dialect.Get(dialect.DriverPostgres)
	oracleD, _ := dialect.Get(dialect.DriverOracle)

	sqliteErr := errors.New("UNIQUE constraint failed: items.tenant_id, items.sku")
	if !sqliteD.IsUniqueConstraintViolation(sqliteErr) {
		t.Error("expected SQLite dialect to recognize unique constraint violation")
	}

	pgErr := errors.New("ERROR: duplicate key value violates unique constraint \"uq_items_tenant_sku\" (SQLSTATE 23505)")
	if !pgD.IsUniqueConstraintViolation(pgErr) {
		t.Error("expected PostgreSQL dialect to recognize unique constraint violation")
	}

	oracleErr := errors.New("ORA-00001: unique constraint (APP.UQ_ITEMS_TENANT_SKU) violated")
	if !oracleD.IsUniqueConstraintViolation(oracleErr) {
		t.Error("expected Oracle dialect to recognize unique constraint violation")
	}

	randomErr := errors.New("connection reset by peer")
	if sqliteD.IsUniqueConstraintViolation(randomErr) || pgD.IsUniqueConstraintViolation(randomErr) || oracleD.IsUniqueConstraintViolation(randomErr) {
		t.Error("random error should not be classified as unique constraint violation")
	}
}

func TestDialect_TimeAndBoolParsing(t *testing.T) {
	sqliteD, _ := dialect.Get(dialect.DriverSQLite)
	pgD, _ := dialect.Get(dialect.DriverPostgres)

	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	timeStr := now.Format(time.RFC3339Nano)

	parsed1, err := sqliteD.ParseTime(timeStr)
	if err != nil || !parsed1.Equal(now) {
		t.Errorf("failed to parse string time in sqlite dialect: %v, %v", parsed1, err)
	}

	parsed2, err := pgD.ParseTime(now)
	if err != nil || !parsed2.Equal(now) {
		t.Errorf("failed to parse time.Time in pg dialect: %v, %v", parsed2, err)
	}

	b1, err := sqliteD.ParseBool(1)
	if err != nil || !b1 {
		t.Errorf("failed to parse integer 1 as bool: %v, %v", b1, err)
	}

	b2, err := pgD.ParseBool(true)
	if err != nil || !b2 {
		t.Errorf("failed to parse native bool: %v, %v", b2, err)
	}
}
