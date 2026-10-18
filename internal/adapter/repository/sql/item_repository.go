package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"new-inventory/internal/adapter/repository/dialect"
	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.ItemRepository = (*ItemRepository)(nil)

type ItemRepository struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewItemRepository(db *sql.DB, d dialect.Dialect) *ItemRepository {
	return &ItemRepository{
		db:      db,
		dialect: d,
	}
}

func (r *ItemRepository) Create(ctx context.Context, item *domain.Item) error {
	query := `
		INSERT INTO items (
			id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		r.dialect.Rebind(query),
		item.ID,
		item.TenantID,
		item.SKU,
		item.Name,
		item.Description,
		item.UnitPrice,
		item.MinThreshold,
		r.dialect.FormatBool(item.IsActive),
		r.dialect.FormatTime(item.CreatedAt),
		r.dialect.FormatTime(item.UpdatedAt),
	)
	if err != nil {
		if r.dialect.IsUniqueConstraintViolation(err) {
			return domain.ErrItemAlreadyExists
		}
		return fmt.Errorf("failed to insert item: %w", err)
	}

	return nil
}

func (r *ItemRepository) Update(ctx context.Context, item *domain.Item) error {
	query := `
		UPDATE items
		SET name = ?, description = ?, unit_price = ?, min_threshold = ?, is_active = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	res, err := r.db.ExecContext(
		ctx,
		r.dialect.Rebind(query),
		item.Name,
		item.Description,
		item.UnitPrice,
		item.MinThreshold,
		r.dialect.FormatBool(item.IsActive),
		r.dialect.FormatTime(item.UpdatedAt),
		item.TenantID,
		item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrItemNotFound
	}

	return nil
}

func (r *ItemRepository) GetBySKU(ctx context.Context, tenantID, sku string) (*domain.Item, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		FROM items
		WHERE tenant_id = ? AND sku = ?
	`
	row := r.db.QueryRowContext(ctx, r.dialect.Rebind(query), tenantID, sku)
	return r.scanItem(row)
}

func (r *ItemRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Item, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		FROM items
		WHERE tenant_id = ? AND id = ?
	`
	row := r.db.QueryRowContext(ctx, r.dialect.Rebind(query), tenantID, id)
	return r.scanItem(row)
}

func (r *ItemRepository) List(ctx context.Context, tenantID string) ([]*domain.Item, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		FROM items
		WHERE tenant_id = ?
		ORDER BY created_at DESC, id ASC
	`
	rows, err := r.db.QueryContext(ctx, r.dialect.Rebind(query), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*domain.Item
	for rows.Next() {
		item, err := r.scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return items, nil
}

func (r *ItemRepository) ListLowStock(ctx context.Context, tenantID string) ([]*domain.LowStockItem, error) {
	query := `
		SELECT
			i.id, i.tenant_id, i.sku, i.name, i.description, i.unit_price, i.min_threshold, i.is_active, i.created_at, i.updated_at,
			COALESCE(SUM(s.quantity), 0) as current_quantity
		FROM items i
		LEFT JOIN stock_levels s ON s.tenant_id = i.tenant_id AND s.item_id = i.id
		WHERE i.tenant_id = ? AND i.is_active = ?
		GROUP BY i.id, i.tenant_id, i.sku, i.name, i.description, i.unit_price, i.min_threshold, i.is_active, i.created_at, i.updated_at
		HAVING COALESCE(SUM(s.quantity), 0) <= i.min_threshold
		ORDER BY current_quantity ASC, i.id ASC
	`
	rows, err := r.db.QueryContext(ctx, r.dialect.Rebind(query), tenantID, r.dialect.FormatBool(true))
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock items: %w", err)
	}
	defer rows.Close()

	var results []*domain.LowStockItem
	for rows.Next() {
		var item domain.Item
		var rawIsActive any
		var rawCreatedAt, rawUpdatedAt any
		var currentQty int

		err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.SKU,
			&item.Name,
			&item.Description,
			&item.UnitPrice,
			&item.MinThreshold,
			&rawIsActive,
			&rawCreatedAt,
			&rawUpdatedAt,
			&currentQty,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan low stock item: %w", err)
		}

		isActive, err := r.dialect.ParseBool(rawIsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item is_active: %w", err)
		}
		item.IsActive = isActive

		createdAt, err := r.dialect.ParseTime(rawCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item created_at: %w", err)
		}
		item.CreatedAt = createdAt

		updatedAt, err := r.dialect.ParseTime(rawUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item updated_at: %w", err)
		}
		item.UpdatedAt = updatedAt

		results = append(results, &domain.LowStockItem{
			Item:            &item,
			CurrentQuantity: currentQty,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("low stock row iteration error: %w", err)
	}

	return results, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *ItemRepository) scanItem(s rowScanner) (*domain.Item, error) {
	var item domain.Item
	var rawIsActive any
	var rawCreatedAt, rawUpdatedAt any

	err := s.Scan(
		&item.ID,
		&item.TenantID,
		&item.SKU,
		&item.Name,
		&item.Description,
		&item.UnitPrice,
		&item.MinThreshold,
		&rawIsActive,
		&rawCreatedAt,
		&rawUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}

	isActive, err := r.dialect.ParseBool(rawIsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item is_active: %w", err)
	}
	item.IsActive = isActive

	createdAt, err := r.dialect.ParseTime(rawCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item created_at: %w", err)
	}
	item.CreatedAt = createdAt

	updatedAt, err := r.dialect.ParseTime(rawUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item updated_at: %w", err)
	}
	item.UpdatedAt = updatedAt

	return &item, nil
}
