package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"new-inventory/internal/domain"
	"new-inventory/internal/port"

	"github.com/google/uuid"
)

var _ port.StockRepository = (*StockRepository)(nil)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) GetStock(ctx context.Context, tenantID, itemID string) (*domain.StockLevel, error) {
	query := `
		SELECT id, tenant_id, item_id, quantity, reserved_quantity, updated_at
		FROM stock_levels
		WHERE tenant_id = ? AND item_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, itemID)

	var stock domain.StockLevel
	var updatedAtStr string
	err := row.Scan(
		&stock.ID,
		&stock.TenantID,
		&stock.ItemID,
		&stock.Quantity,
		&stock.ReservedQuantity,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStockNotFound
		}
		return nil, fmt.Errorf("failed to query stock level: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err == nil {
		stock.UpdatedAt = updatedAt
	}

	return &stock, nil
}

func (r *StockRepository) RecordMovement(ctx context.Context, movement *domain.StockMovement, newBalance int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Insert stock movement audit log
	insertMovement := `
		INSERT INTO stock_movements (
			id, tenant_id, item_id, movement_type, quantity, balance_after, reference, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(
		ctx,
		insertMovement,
		movement.ID,
		movement.TenantID,
		movement.ItemID,
		movement.MovementType,
		movement.Quantity,
		movement.BalanceAfter,
		movement.Reference,
		movement.Notes,
		movement.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock movement: %w", err)
	}

	// 2. Upsert stock level
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	upsertStock := `
		INSERT INTO stock_levels (id, tenant_id, item_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, 0, ?)
		ON CONFLICT(tenant_id, item_id) DO UPDATE SET
			quantity = excluded.quantity,
			updated_at = excluded.updated_at
	`
	stockLevelID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		upsertStock,
		stockLevelID,
		movement.TenantID,
		movement.ItemID,
		newBalance,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert stock level: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit stock transaction: %w", err)
	}

	return nil
}

func (r *StockRepository) ListMovements(ctx context.Context, tenantID, itemID string) ([]*domain.StockMovement, error) {
	query := `
		SELECT id, tenant_id, item_id, movement_type, quantity, balance_after, reference, notes, created_at
		FROM stock_movements
		WHERE tenant_id = ? AND item_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		var createdAtStr string
		err := rows.Scan(
			&m.ID,
			&m.TenantID,
			&m.ItemID,
			&m.MovementType,
			&m.Quantity,
			&m.BalanceAfter,
			&m.Reference,
			&m.Notes,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock movement: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err == nil {
			m.CreatedAt = createdAt
		}
		movements = append(movements, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("movement row iteration error: %w", err)
	}

	return movements, nil
}
