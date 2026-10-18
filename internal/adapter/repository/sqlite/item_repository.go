package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"
)

var _ port.ItemRepository = (*ItemRepository)(nil)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, item *domain.Item) error {
	query := `
		INSERT INTO items (
			id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isActiveInt := 0
	if item.IsActive {
		isActiveInt = 1
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.TenantID,
		item.SKU,
		item.Name,
		item.Description,
		item.UnitPrice,
		item.MinThreshold,
		isActiveInt,
		item.CreatedAt.Format(time.RFC3339Nano),
		item.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
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

	isActiveInt := 0
	if item.IsActive {
		isActiveInt = 1
	}

	res, err := r.db.ExecContext(
		ctx,
		query,
		item.Name,
		item.Description,
		item.UnitPrice,
		item.MinThreshold,
		isActiveInt,
		item.UpdatedAt.Format(time.RFC3339Nano),
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
	row := r.db.QueryRowContext(ctx, query, tenantID, sku)
	return scanItem(row)
}

func (r *ItemRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Item, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		FROM items
		WHERE tenant_id = ? AND id = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)
	return scanItem(row)
}

func (r *ItemRepository) List(ctx context.Context, tenantID string) ([]*domain.Item, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit_price, min_threshold, is_active, created_at, updated_at
		FROM items
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*domain.Item
	for rows.Next() {
		item, err := scanItem(rows)
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
			COALESCE(s.quantity, 0) as current_quantity
		FROM items i
		LEFT JOIN stock_levels s ON s.tenant_id = i.tenant_id AND s.item_id = i.id
		WHERE i.tenant_id = ? AND i.is_active = 1 AND COALESCE(s.quantity, 0) <= i.min_threshold
		ORDER BY COALESCE(s.quantity, 0) ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock items: %w", err)
	}
	defer rows.Close()

	var results []*domain.LowStockItem
	for rows.Next() {
		var item domain.Item
		var isActiveInt int
		var createdAtStr, updatedAtStr string
		var currentQty int

		err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.SKU,
			&item.Name,
			&item.Description,
			&item.UnitPrice,
			&item.MinThreshold,
			&isActiveInt,
			&createdAtStr,
			&updatedAtStr,
			&currentQty,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan low stock item: %w", err)
		}

		item.IsActive = isActiveInt == 1
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err == nil {
			item.CreatedAt = createdAt
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
		if err == nil {
			item.UpdatedAt = updatedAt
		}

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

func scanItem(s rowScanner) (*domain.Item, error) {
	var item domain.Item
	var isActiveInt int
	var createdAtStr, updatedAtStr string

	err := s.Scan(
		&item.ID,
		&item.TenantID,
		&item.SKU,
		&item.Name,
		&item.Description,
		&item.UnitPrice,
		&item.MinThreshold,
		&isActiveInt,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}

	item.IsActive = isActiveInt == 1
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err == nil {
		item.CreatedAt = createdAt
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err == nil {
		item.UpdatedAt = updatedAt
	}

	return &item, nil
}
