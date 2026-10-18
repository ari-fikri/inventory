package sqlite

import (
	"database/sql"

	"new-inventory/internal/adapter/repository/dialect"
	reposql "new-inventory/internal/adapter/repository/sql"
)

type SupplierRepository = reposql.SupplierRepository

func NewSupplierRepository(db *sql.DB) *reposql.SupplierRepository {
	return reposql.NewSupplierRepository(db, dialect.NewSQLiteDialect())
}

type PurchaseOrderRepository = reposql.PurchaseOrderRepository

func NewPurchaseOrderRepository(db *sql.DB) *reposql.PurchaseOrderRepository {
	return reposql.NewPurchaseOrderRepository(db, dialect.NewSQLiteDialect())
}
