package dialect

import (
	"strings"
	"time"
)

type PostgresDialect struct{}

var _ Dialect = (*PostgresDialect)(nil)

func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{}
}

func (d *PostgresDialect) DriverName() Driver {
	return DriverPostgres
}

func (d *PostgresDialect) PlaceholderStyle() PlaceholderStyle {
	return PlaceholderDollar
}

func (d *PostgresDialect) Rebind(query string) string {
	return RebindQuery(query, PlaceholderDollar)
}

func (d *PostgresDialect) IsUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Standard PostgreSQL SQLSTATE 23505 (unique_violation)
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "unique constraint")
}

func (d *PostgresDialect) UpsertStockLevelQuery() string {
	return `
		INSERT INTO stock_levels (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6)
		ON CONFLICT (tenant_id, item_id, warehouse_id) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			updated_at = EXCLUDED.updated_at
	`
}

func (d *PostgresDialect) UpsertTenantSettingsQuery() string {
	return `
		INSERT INTO tenant_settings (tenant_id, valuation_method, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id) DO UPDATE SET
			valuation_method = EXCLUDED.valuation_method,
			updated_at = EXCLUDED.updated_at
	`
}

func (d *PostgresDialect) FormatTime(t time.Time) any {
	return t.UTC()
}

func (d *PostgresDialect) ParseTime(val any) (time.Time, error) {
	return CommonParseTime(val)
}

func (d *PostgresDialect) FormatBool(b bool) any {
	return b
}

func (d *PostgresDialect) ParseBool(val any) (bool, error) {
	return CommonParseBool(val)
}
