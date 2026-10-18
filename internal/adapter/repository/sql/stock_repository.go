package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"new-inventory/internal/adapter/repository/dialect"
	"new-inventory/internal/domain"
	"new-inventory/internal/port"

	"github.com/google/uuid"
)

var _ port.StockRepository = (*StockRepository)(nil)

type StockRepository struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewStockRepository(db *sql.DB, d dialect.Dialect) *StockRepository {
	return &StockRepository{
		db:      db,
		dialect: d,
	}
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
	row := r.db.QueryRowContext(ctx, r.dialect.Rebind(query), tenantID, itemID, warehouseID)

	var stock domain.StockLevel
	var rawUpdatedAt any
	err := row.Scan(
		&stock.ID,
		&stock.TenantID,
		&stock.ItemID,
		&stock.WarehouseID,
		&stock.Quantity,
		&stock.ReservedQuantity,
		&rawUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStockNotFound
		}
		return nil, fmt.Errorf("failed to query stock level: %w", err)
	}

	updatedAt, err := r.dialect.ParseTime(rawUpdatedAt)
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
		r.dialect.Rebind(insertMovement),
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
		r.dialect.FormatTime(movement.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock movement: %w", err)
	}

	// 2. Upsert stock level using dialect-specific query
	now := r.dialect.FormatTime(time.Now().UTC())
	stockLevelID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		r.dialect.UpsertStockLevelQuery(),
		stockLevelID,
		movement.TenantID,
		movement.ItemID,
		movement.WarehouseID,
		newBalance,
		now,
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
			r.dialect.Rebind(insertCostLayer),
			costLayer.ID,
			costLayer.TenantID,
			costLayer.ItemID,
			costLayer.WarehouseID,
			costLayer.Quantity,
			costLayer.RemainingQuantity,
			costLayer.UnitCost,
			r.dialect.FormatTime(costLayer.CreatedAt),
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
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.db.QueryContext(ctx, r.dialect.Rebind(query), tenantID, itemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active cost layers: %w", err)
	}
	defer rows.Close()

	var layers []domain.CostLayer
	for rows.Next() {
		var l domain.CostLayer
		var rawCreatedAt any
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ItemID, &l.WarehouseID, &l.Quantity, &l.RemainingQuantity, &l.UnitCost, &rawCreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan active cost layer: %w", err)
		}
		if t, err := r.dialect.ParseTime(rawCreatedAt); err == nil {
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
		r.dialect.Rebind(insertMovement),
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
		r.dialect.FormatTime(movement.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock movement: %w", err)
	}

	now := r.dialect.FormatTime(time.Now().UTC())
	stockLevelID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		r.dialect.UpsertStockLevelQuery(),
		stockLevelID,
		movement.TenantID,
		movement.ItemID,
		movement.WarehouseID,
		newBalance,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert stock level: %w", err)
	}

	updateLayerQuery := r.dialect.Rebind(`UPDATE stock_cost_layers SET remaining_quantity = ? WHERE id = ?`)
	for _, c := range consumed {
		_, err := tx.ExecContext(ctx, updateLayerQuery, c.RemainingQty, c.LayerID)
		if err != nil {
			return fmt.Errorf("failed to update consumed cost layer %s: %w", c.LayerID, err)
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
			id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, total_cogs, reference, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	reboundInsert := r.dialect.Rebind(insertMovement)

	// 1. Insert OUT movement
	_, err = tx.ExecContext(
		ctx,
		reboundInsert,
		outMovement.ID,
		outMovement.TenantID,
		outMovement.ItemID,
		outMovement.WarehouseID,
		outMovement.MovementType,
		outMovement.Quantity,
		outMovement.BalanceAfter,
		outMovement.TotalCOGS,
		outMovement.Reference,
		outMovement.Notes,
		r.dialect.FormatTime(outMovement.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("failed to insert transfer-out movement: %w", err)
	}

	// 2. Insert IN movement
	_, err = tx.ExecContext(
		ctx,
		reboundInsert,
		inMovement.ID,
		inMovement.TenantID,
		inMovement.ItemID,
		inMovement.WarehouseID,
		inMovement.MovementType,
		inMovement.Quantity,
		inMovement.BalanceAfter,
		inMovement.TotalCOGS,
		inMovement.Reference,
		inMovement.Notes,
		r.dialect.FormatTime(inMovement.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("failed to insert transfer-in movement: %w", err)
	}

	// 3. Upsert source warehouse stock level
	now := r.dialect.FormatTime(time.Now().UTC())
	sourceLevelID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		r.dialect.UpsertStockLevelQuery(),
		sourceLevelID,
		outMovement.TenantID,
		outMovement.ItemID,
		outMovement.WarehouseID,
		newSourceBal,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert source stock level: %w", err)
	}

	// 4. Upsert target warehouse stock level
	targetLevelID := uuid.NewString()
	_, err = tx.ExecContext(
		ctx,
		r.dialect.UpsertStockLevelQuery(),
		targetLevelID,
		inMovement.TenantID,
		inMovement.ItemID,
		inMovement.WarehouseID,
		newTargetBal,
		now,
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
		SELECT id, tenant_id, item_id, warehouse_id, movement_type, quantity, balance_after, total_cogs, reference, notes, created_at
		FROM stock_movements
		WHERE tenant_id = ? AND item_id = ?
		ORDER BY created_at DESC, id DESC
	`
	rows, err := r.db.QueryContext(ctx, r.dialect.Rebind(query), tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		var rawCreatedAt any
		err := rows.Scan(
			&m.ID,
			&m.TenantID,
			&m.ItemID,
			&m.WarehouseID,
			&m.MovementType,
			&m.Quantity,
			&m.BalanceAfter,
			&m.TotalCOGS,
			&m.Reference,
			&m.Notes,
			&rawCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock movement: %w", err)
		}
		if t, err := r.dialect.ParseTime(rawCreatedAt); err == nil {
			m.CreatedAt = t
		}
		movements = append(movements, &m)
	}

	return movements, rows.Err()
}
