package sqlite

import (
	"database/sql"

	"new-inventory/internal/adapter/repository/dialect"
	reposql "new-inventory/internal/adapter/repository/sql"
)

type ValuationRepository = reposql.ValuationRepository

func NewValuationRepository(db *sql.DB) *reposql.ValuationRepository {
	return reposql.NewValuationRepository(db, dialect.NewSQLiteDialect())
}
