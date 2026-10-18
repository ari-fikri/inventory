package sqlite

import (
	"database/sql"

	"new-inventory/internal/adapter/repository/dialect"
	reposql "new-inventory/internal/adapter/repository/sql"
)

type StockRepository = reposql.StockRepository

func NewStockRepository(db *sql.DB) *reposql.StockRepository {
	return reposql.NewStockRepository(db, dialect.NewSQLiteDialect())
}
