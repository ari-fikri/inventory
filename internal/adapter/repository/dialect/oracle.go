package dialect

import (
	"strings"
	"time"
)

type OracleDialect struct{}

var _ Dialect = (*OracleDialect)(nil)

func NewOracleDialect() *OracleDialect {
	return &OracleDialect{}
}

func (d *OracleDialect) DriverName() Driver {
	return DriverOracle
}

func (d *OracleDialect) PlaceholderStyle() PlaceholderStyle {
	return PlaceholderColon
}

func (d *OracleDialect) Rebind(query string) string {
	return RebindQuery(query, PlaceholderColon)
}

func (d *OracleDialect) IsUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Standard Oracle error ORA-00001 (unique constraint violated)
	return strings.Contains(msg, "ORA-00001") ||
		strings.Contains(msg, "unique constraint")
}

func (d *OracleDialect) UpsertStockLevelQuery() string {
	return `
		MERGE INTO stock_levels dst
		USING (SELECT :1 AS id, :2 AS tenant_id, :3 AS item_id, :4 AS warehouse_id, :5 AS quantity, 0 AS reserved_quantity, :6 AS updated_at FROM dual) src
		ON (dst.tenant_id = src.tenant_id AND dst.item_id = src.item_id AND dst.warehouse_id = src.warehouse_id)
		WHEN MATCHED THEN
			UPDATE SET dst.quantity = src.quantity, dst.updated_at = src.updated_at
		WHEN NOT MATCHED THEN
			INSERT (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
			VALUES (src.id, src.tenant_id, src.item_id, src.warehouse_id, src.quantity, src.reserved_quantity, src.updated_at)
	`
}

func (d *OracleDialect) UpsertTenantSettingsQuery() string {
	return `
		MERGE INTO tenant_settings dst
		USING (SELECT :1 AS tenant_id, :2 AS valuation_method, :3 AS updated_at FROM dual) src
		ON (dst.tenant_id = src.tenant_id)
		WHEN MATCHED THEN
			UPDATE SET dst.valuation_method = src.valuation_method, dst.updated_at = src.updated_at
		WHEN NOT MATCHED THEN
			INSERT (tenant_id, valuation_method, updated_at)
			VALUES (src.tenant_id, src.valuation_method, src.updated_at)
	`
}

func (d *OracleDialect) FormatTime(t time.Time) any {
	return t.UTC()
}

func (d *OracleDialect) ParseTime(val any) (time.Time, error) {
	return CommonParseTime(val)
}

func (d *OracleDialect) FormatBool(b bool) any {
	if b {
		return 1
	}
	return 0
}

func (d *OracleDialect) ParseBool(val any) (bool, error) {
	return CommonParseBool(val)
}
