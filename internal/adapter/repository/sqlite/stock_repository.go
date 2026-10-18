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

func (r *StockRepository) GetStock(ctx context.Context, tenantID, itemID, warehouseID string) (*domain.StockLevel, error) {
	if warehouseID == "" {
		warehouseID = "default"
	}
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at
		FROM stock_levels
		WHERE tenant_id = ? AND item_id = ? AND warehouse_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, itemID, warehouseID)

	var stock domain.StockLevel
	var updatedAtStr string
	err := row.Scan(
		&stock.ID,
		&stock.TenantID,
		&stock.ItemID,
		&stock.WarehouseID,
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
	return r.RecordMovementWithCostLayer(ctx, movement, newBalance, nil)
}

func (r *StockRepository) RecordMovementWithCostLayer(ctx context.Context, movement *domain.StockMovement, newBalance int, costLayer *domain.CostLayer) error {
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
			id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, total_cogs, reference, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(
		ctx,
		insertMovement,
		movement.ID,
		movement.TenantID,
		movement.ItemID,
		movement.WarehouseID,
		movement.MovementType,
		movement.Quantity,
		movement.BalanceAfter,
		movement.TotalCOGS,
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
		INSERT INTO stock_levels (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?)
		ON CONFLICT(tenant_id, item_id, warehouse_id) DO UPDATE SET
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
		movement.WarehouseID,
		newBalance,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert stock level: %w", err)
	}

	// 3. Optional insert cost layer
	if costLayer != nil {
		insertCostLayer := `
			INSERT INTO stock_cost_layers (
				id, tenant_id, item_id, warehouse_id, quantity, remaining_quantity, unit_cost, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err = tx.ExecContext(
			ctx,
			insertCostLayer,
			costLayer.ID,
			costLayer.TenantID,
			costLayer.ItemID,
			costLayer.WarehouseID,
			costLayer.Quantity,
			costLayer.RemainingQuantity,
			costLayer.UnitCost,
			costLayer.CreatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("failed to insert cost layer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit stock transaction: %w", err)
	}

	return nil
}

func (r *StockRepository) GetActiveCostLayers(ctx context.Context, tenantID, itemID, warehouseID string) ([]domain.CostLayer, error) {
	if warehouseID == "" {
		warehouseID = "default"
	}
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, quantity, remaining_quantity, unit_cost, created_at
		FROM stock_cost_layers
		WHERE tenant_id = ? AND item_id = ? AND warehouse_id = ? AND remaining_quantity > 0
		ORDER BY rowid ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, itemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active cost layers: %w", err)
	}
	defer rows.Close()

	var layers []domain.CostLayer
	for rows.Next() {
		var l domain.CostLayer
		var createdAtStr string
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ItemID, &l.WarehouseID, &l.Quantity, &l.RemainingQuantity, &l.UnitCost, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan active cost layer: %w", err)
		}
		if t, err := time.Parse(time.RFC3339Nano, createdAtStr); err == nil {
			l.CreatedAt = t
		}
		layers = append(layers, l)
	}
	return layers, rows.Err()
}

func (r *StockRepository) RecordStockOutWithValuation(ctx context.Context, movement *domain.StockMovement, newBalance int, consumed []domain.ConsumedCostLayer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertMovement := `
		INSERT INTO stock_movements (
			id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, total_cogs, reference, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(
		ctx,
		insertMovement,
		movement.ID,
		movement.TenantID,
		movement.ItemID,
		movement.WarehouseID,
		movement.MovementType,
		movement.Quantity,
		movement.BalanceAfter,
		movement.TotalCOGS,
		movement.Reference,
		movement.Notes,
		movement.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock movement: %w", err)
	}

	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	upsertStock := `
		INSERT INTO stock_levels (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?)
		ON CONFLICT(tenant_id, item_id, warehouse_id) DO UPDATE SET
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
		movement.WarehouseID,
		newBalance,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert stock level: %w", err)
	}

	updateLayer := `UPDATE stock_cost_layers SET remaining_quantity = ? WHERE id = ?`
	for _, c := range consumed {
		_, err = tx.ExecContext(ctx, updateLayer, c.RemainingQty, c.LayerID)
		if err != nil {
			return fmt.Errorf("failed to update cost layer %s: %w", c.LayerID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit stock out transaction: %w", err)
	}

	return nil
}

func (r *StockRepository) RecordTransfer(ctx context.Context, outMovement, inMovement *domain.StockMovement, newSourceBal, newTargetBal int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertMovement := `
		INSERT INTO stock_movements (
			id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, reference, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.ExecContext(
		ctx,
		insertMovement,
		outMovement.ID,
		outMovement.TenantID,
		outMovement.ItemID,
		outMovement.WarehouseID,
		outMovement.MovementType,
		outMovement.Quantity,
		outMovement.BalanceAfter,
		outMovement.Reference,
		outMovement.Notes,
		outMovement.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to insert transfer out movement: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		insertMovement,
		inMovement.ID,
		inMovement.TenantID,
		inMovement.ItemID,
		inMovement.WarehouseID,
		inMovement.MovementType,
		inMovement.Quantity,
		inMovement.BalanceAfter,
		inMovement.Reference,
		inMovement.Notes,
		inMovement.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to insert transfer in movement: %w", err)
	}

	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	upsertStock := `
		INSERT INTO stock_levels (id, tenant_id, item_id, warehouse_id, quantity, reserved_quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?)
		ON CONFLICT(tenant_id, item_id, warehouse_id) DO UPDATE SET
			quantity = excluded.quantity,
			updated_at = excluded.updated_at
	`

	_, err = tx.ExecContext(
		ctx,
		upsertStock,
		uuid.NewString(),
		outMovement.TenantID,
		outMovement.ItemID,
		outMovement.WarehouseID,
		newSourceBal,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert source stock level: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		upsertStock,
		uuid.NewString(),
		inMovement.TenantID,
		inMovement.ItemID,
		inMovement.WarehouseID,
		newTargetBal,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert target stock level: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transfer transaction: %w", err)
	}

	return nil
}

func (r *StockRepository) ListMovements(ctx context.Context, tenantID, itemID string) ([]*domain.StockMovement, error) {
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, reference, notes, created_at
		FROM stock_movements
		WHERE tenant_id = ? AND item_id = ?
		ORDER BY rowid DESC
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
			&m.WarehouseID,
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
