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
)

var _ port.ValuationRepository = (*ValuationRepository)(nil)

type ValuationRepository struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewValuationRepository(db *sql.DB, d dialect.Dialect) *ValuationRepository {
	return &ValuationRepository{
		db:      db,
		dialect: d,
	}
}

func (r *ValuationRepository) GetCostLayers(ctx context.Context, tenantID, itemID string) ([]domain.CostLayer, error) {
	query := `
		SELECT id, tenant_id, item_id, warehouse_id, quantity, remaining_quantity, unit_cost, created_at
		FROM stock_cost_layers
		WHERE tenant_id = ? AND item_id = ?
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.db.QueryContext(ctx, r.dialect.Rebind(query), tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost layers: %w", err)
	}
	defer rows.Close()

	var layers []domain.CostLayer
	for rows.Next() {
		var l domain.CostLayer
		var rawCreatedAt any
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ItemID, &l.WarehouseID, &l.Quantity, &l.RemainingQuantity, &l.UnitCost, &rawCreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cost layer: %w", err)
		}
		if t, err := r.dialect.ParseTime(rawCreatedAt); err == nil {
			l.CreatedAt = t
		}
		layers = append(layers, l)
	}
	return layers, rows.Err()
}

func (r *ValuationRepository) GetActiveCostLayers(ctx context.Context, tenantID, itemID, warehouseID string) ([]domain.CostLayer, error) {
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

func (r *ValuationRepository) UpdateCostLayerRemaining(ctx context.Context, layerID string, remainingQty int) error {
	query := `UPDATE stock_cost_layers SET remaining_quantity = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Rebind(query), remainingQty, layerID)
	if err != nil {
		return fmt.Errorf("failed to update cost layer remaining quantity: %w", err)
	}
	return nil
}

func (r *ValuationRepository) GetTenantValuationMethod(ctx context.Context, tenantID string) (string, error) {
	query := `SELECT valuation_method FROM tenant_settings WHERE tenant_id = ?`
	row := r.db.QueryRowContext(ctx, r.dialect.Rebind(query), tenantID)
	var method string
	if err := row.Scan(&method); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ValuationMethodFIFO, nil
		}
		return "", fmt.Errorf("failed to query tenant valuation method: %w", err)
	}
	return method, nil
}

func (r *ValuationRepository) SetTenantValuationMethod(ctx context.Context, tenantID, method string) error {
	now := r.dialect.FormatTime(time.Now().UTC())
	_, err := r.db.ExecContext(
		ctx,
		r.dialect.UpsertTenantSettingsQuery(),
		tenantID,
		method,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert tenant valuation method: %w", err)
	}
	return nil
}

func (r *ValuationRepository) GetTenantTotalStock(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COALESCE(SUM(quantity), 0) FROM stock_levels WHERE tenant_id = ?`
	row := r.db.QueryRowContext(ctx, r.dialect.Rebind(query), tenantID)
	var total int
	if err := row.Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to calculate tenant total stock: %w", err)
	}
	return total, nil
}
