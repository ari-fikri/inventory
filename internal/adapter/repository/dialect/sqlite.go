package dialect

import (
	"strings"
	"time"
)

type SQLiteDialect struct{}

var _ Dialect = (*SQLiteDialect)(nil)

func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{}
}

func (d *SQLiteDialect) DriverName() Driver {
	return DriverSQLite
}

func (d *SQLiteDialect) PlaceholderStyle() PlaceholderStyle {
	return PlaceholderQuestion
}

func (d *SQLiteDialect) Rebind(query string) string {
	return query
}

func (d *SQLiteDialect) IsUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "PRIMARY KEY must be unique") ||
		strings.Contains(msg, "constraint failed: UNIQUE")
}

func (d *SQLiteDialect) UpsertStockLevelQuery() string {
	return `
		INSERT INTO stock_levels (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?)
		ON CONFLICT(tenant_id, item_id, warehouse_id) DO UPDATE SET
			quantity = excluded.quantity,
			updated_at = excluded.updated_at
	`
}

func (d *SQLiteDialect) UpsertTenantSettingsQuery() string {
	return `
		INSERT INTO tenant_settings (tenant_id, valuation_method, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(tenant_id) DO UPDATE SET
			valuation_method = excluded.valuation_method,
			updated_at = excluded.updated_at
	`
}

func (d *SQLiteDialect) FormatTime(t time.Time) any {
	return t.UTC().Format(time.RFC3339Nano)
}

func (d *SQLiteDialect) ParseTime(val any) (time.Time, error) {
	return CommonParseTime(val)
}

func (d *SQLiteDialect) FormatBool(b bool) any {
	if b {
		return 1
	}
	return 0
}

func (d *SQLiteDialect) ParseBool(val any) (bool, error) {
	return CommonParseBool(val)
}
