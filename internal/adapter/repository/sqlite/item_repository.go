package sqlite

import (
	"database/sql"

	"new-inventory/internal/adapter/repository/dialect"
	reposql "new-inventory/internal/adapter/repository/sql"
)

type ItemRepository = reposql.ItemRepository

func NewItemRepository(db *sql.DB) *reposql.ItemRepository {
	return reposql.NewItemRepository(db, dialect.NewSQLiteDialect())
}
